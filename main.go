package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const (
	// 默认输入和输出目录
	defaultInputDir  = "input"
	defaultOutputDir = "output"
)

// 默认工作协程数
var defaultWorkerCount = runtime.NumCPU()

// 支持的视频格式
var supportedFormats = []string{".mp4", ".avi", ".mov", ".mkv", ".flv", ".wmv", ".m4v", ".3gp", ".webm"}

// ==================== 接口定义 ====================

// VideoProcessor 视频处理器接口
type VideoProcessor interface {
	ProcessAllVideos() error
	ProcessVideo(inputPath string) error
}

// ConfigProvider 配置提供者接口
type ConfigProvider interface {
	GetConfigs() []VideoConfig
	ValidateConfig(config VideoConfig) error
}

// PathGenerator 路径生成器接口
type PathGenerator interface {
	GenerateOutputPath(inputPath string, config VideoConfig) (string, error)
	EnsureDir(dir string) error
}

// VideoConverter 视频转换器接口
type VideoConverter interface {
	ConvertVideo(inputPath, outputPath string, config VideoConfig, audioConfig *Config) error
	GetVideoDuration(videoPath string) (float64, error)
}

// ClipCalculator 剪辑计算器接口
type ClipCalculator interface {
	CalculateClipTimes(totalDuration float64, config VideoConfig) (float64, float64, error)
}

// ==================== 数据结构 ====================

// VideoConfig 视频配置
type VideoConfig struct {
	Width        int     `json:"Width"`        // 视频宽度
	Height       int     `json:"Height"`       // 视频高度
	ClipDuration int     `json:"ClipDuration"` // 剪辑时长（秒）
	Speed        float64 `json:"Speed"`        // 播放速度倍数
	VideoBitrate int     `json:"VideoBitrate"` // 视频比特率(kbps)
	ClipStrategy string  `json:"ClipStrategy"` // 剪辑策略：start_segments 或 end_segments
	OutputSuffix string  `json:"OutputSuffix"` // 输出文件后缀
	OutputFolder string  `json:"OutputFolder"` // 输出文件夹
}

// ==================== 接口实现 ====================

// ConfigurableConfigProvider 可配置的配置提供者
type ConfigurableConfigProvider struct {
	config *Config
}

func NewConfigurableConfigProvider(configPath string) (ConfigProvider, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	return &ConfigurableConfigProvider{config: config}, nil
}

func (ccp *ConfigurableConfigProvider) GetConfigs() []VideoConfig {
	return ccp.config.VideoConfigs
}

func (ccp *ConfigurableConfigProvider) ValidateConfig(config VideoConfig) error {
	if config.Width <= 0 || config.Height <= 0 {
		return fmt.Errorf("invalid resolution: %dx%d", config.Width, config.Height)
	}
	if config.Speed <= 0 {
		return fmt.Errorf("invalid speed: %.1f", config.Speed)
	}
	if config.ClipDuration <= 0 {
		return fmt.Errorf("invalid clip duration: %d", config.ClipDuration)
	}
	return nil
}

// DefaultConfigProvider 默认配置提供者（保持向后兼容）
type DefaultConfigProvider struct{}

func NewDefaultConfigProvider() ConfigProvider {
	return &DefaultConfigProvider{}
}

func (dcp *DefaultConfigProvider) GetConfigs() []VideoConfig {
	defaultConfig := GetDefaultConfig()
	return defaultConfig.VideoConfigs
}

func (dcp *DefaultConfigProvider) ValidateConfig(config VideoConfig) error {
	if config.Width <= 0 || config.Height <= 0 {
		return fmt.Errorf("invalid resolution: %dx%d", config.Width, config.Height)
	}
	if config.Speed <= 0 {
		return fmt.Errorf("invalid speed: %.1f", config.Speed)
	}
	if config.ClipDuration <= 0 {
		return fmt.Errorf("invalid clip duration: %d", config.ClipDuration)
	}
	return nil
}

// DefaultPathGenerator 默认路径生成器
type DefaultPathGenerator struct {
	inputDir  string
	outputDir string
}

func NewDefaultPathGenerator(inputDir, outputDir string) PathGenerator {
	return &DefaultPathGenerator{
		inputDir:  inputDir,
		outputDir: outputDir,
	}
}

func (dpg *DefaultPathGenerator) GenerateOutputPath(inputPath string, config VideoConfig) (string, error) {
	relPath, err := filepath.Rel(dpg.inputDir, inputPath)
	if err != nil {
		return "", fmt.Errorf("计算相对路径失败: %v", err)
	}

	dir := filepath.Dir(relPath)
	filename := filepath.Base(relPath)
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	// 生成新文件名：原名 + 后缀 + 扩展名
	newFilename := fmt.Sprintf("%s%s%s", name, config.OutputSuffix, ext)

	// 生成完整输出路径
	var outputPath string
	if dir == "." {
		outputPath = filepath.Join(dpg.outputDir, config.OutputFolder, newFilename)
	} else {
		outputPath = filepath.Join(dpg.outputDir, config.OutputFolder, dir, newFilename)
	}

	return outputPath, nil
}

func (dpg *DefaultPathGenerator) EnsureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// DefaultClipCalculator 默认剪辑计算器
type DefaultClipCalculator struct{}

func NewDefaultClipCalculator() ClipCalculator {
	return &DefaultClipCalculator{}
}

func (dcc *DefaultClipCalculator) CalculateClipTimes(totalDuration float64, config VideoConfig) (float64, float64, error) {
	var startTime, endTime float64

	switch config.ClipStrategy {
	case "start_segments":
		// 从开始第5秒开始截取20秒 (5s-25s)
		startTime = 5.0
		endTime = startTime + float64(config.ClipDuration)
		if endTime > totalDuration {
			endTime = totalDuration
		}
	case "end_segments":
		// 从倒数第25秒开始截取20秒 (倒数25s-倒数5s)
		endTime = totalDuration - 5
		startTime = endTime - float64(config.ClipDuration)
		if startTime < 0 {
			startTime = 0
		}
	case "last_segments":
		// 截取倒数第5秒的前N秒
		endTime = totalDuration - 5
		startTime = endTime - float64(config.ClipDuration)
		if startTime < 0 {
			startTime = 0
		}
	case "middle_segments":
		// 截取开始前5秒至结束前5秒之间的N秒内容
		availableStart := 5.0               // 开始前5秒后
		availableEnd := totalDuration - 5.0 // 结束前5秒前
		availableDuration := availableEnd - availableStart

		if availableDuration < float64(config.ClipDuration) {
			// 如果可用时长不足，取全部可用时长
			startTime = availableStart
			endTime = availableEnd
		} else {
			// 从可用时长中间取指定时长
			middlePoint := availableStart + availableDuration/2
			startTime = middlePoint - float64(config.ClipDuration)/2
			endTime = startTime + float64(config.ClipDuration)
		}
	default:
		return 0, 0, fmt.Errorf("不支持的剪辑策略: %s", config.ClipStrategy)
	}

	return startTime, endTime, nil
}

// FFmpegVideoConverter FFmpeg视频转换器
type FFmpegVideoConverter struct{}

func NewFFmpegVideoConverter() VideoConverter {
	return &FFmpegVideoConverter{}
}

func (fvc *FFmpegVideoConverter) GetVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "csv=p=0", videoPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("无法获取视频时长: %v", err)
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("解析视频时长失败: %v", err)
	}

	return duration, nil
}

func (fvc *FFmpegVideoConverter) ConvertVideo(inputPath, outputPath string, config VideoConfig, audioConfig *Config) error {
	// 第一步：剪辑和基础转换
	tempPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + "_temp" + filepath.Ext(outputPath)
	if err := fvc.clipAndConvert(inputPath, tempPath, config, audioConfig); err != nil {
		return fmt.Errorf("剪辑转换失败: %v", err)
	}

	// 第二步：应用速度效果和最终压缩
	if err := fvc.applySpeedAndCompress(tempPath, outputPath, config, audioConfig); err != nil {
		_ = os.Remove(tempPath) // 清理临时文件
		return fmt.Errorf("应用速度效果失败: %v", err)
	}

	// 清理临时文件
	_ = os.Remove(tempPath)

	return nil
}

func (fvc *FFmpegVideoConverter) clipAndConvert(inputPath, outputPath string, config VideoConfig, audioConfig *Config) error {
	args := []string{"-i", inputPath}

	// 生成视频滤镜
	var videoFilter string
	if config.Width == config.Height {
		// 正方形：缩放并裁剪
		videoFilter = fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d",
			config.Width, config.Height, config.Width, config.Height)
	} else {
		// 矩形：直接缩放
		videoFilter = fmt.Sprintf("scale=%d:%d", config.Width, config.Height)
	}

	// Adobe Media Encoder风格的VBR编码参数
	args = append(args,
		"-vf", videoFilter,
		"-c:v", "libx264",
		"-preset", "medium",
		"-profile:v", "high",
		"-level", "4.1",
		"-b:v", fmt.Sprintf("%dk", config.VideoBitrate),
		"-maxrate", fmt.Sprintf("%dk", int(float64(config.VideoBitrate)*1.2)),
		"-bufsize", fmt.Sprintf("%dk", config.VideoBitrate*2),
		"-g", "50",
		"-keyint_min", "25",
		"-sc_threshold", "40",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", audioConfig.AudioBitrate,
		"-ar", "48000",
		"-y", outputPath,
	)

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg剪辑失败: %v\n输出: %s", err, string(output))
	}

	return nil
}

func (fvc *FFmpegVideoConverter) applySpeedAndCompress(inputPath, outputPath string, config VideoConfig, audioConfig *Config) error {
	speedPTS := 1.0 / config.Speed // 视频PTS调整

	// Adobe Media Encoder风格的高质量VBR编码
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-filter:v", fmt.Sprintf("setpts=%.3f*PTS", speedPTS),
		"-filter:a", fmt.Sprintf("atempo=%.1f", config.Speed),
		"-c:v", "libx264",
		"-preset", "slow",
		"-profile:v", "high",
		"-level", "4.1",
		"-b:v", fmt.Sprintf("%dk", config.VideoBitrate),
		"-maxrate", fmt.Sprintf("%dk", int(float64(config.VideoBitrate)*1.25)),
		"-bufsize", fmt.Sprintf("%dk", config.VideoBitrate*2),
		"-g", "50",
		"-keyint_min", "25",
		"-sc_threshold", "40",
		"-bf", "3",
		"-b_strategy", "2",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", audioConfig.AudioBitrate,
		"-ar", "48000",
		"-movflags", "+faststart",
		"-y", outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("应用速度效果失败: %v\n输出: %s", err, string(output))
	}

	return nil
}

// DefaultVideoProcessor 默认视频处理器
type DefaultVideoProcessor struct {
	inputDir       string
	outputDir      string
	config         *Config
	configProvider ConfigProvider
	pathGenerator  PathGenerator
	converter      VideoConverter
	clipCalculator ClipCalculator
}

func NewVideoProcessor(inputDir, outputDir string, config *Config, configProvider ConfigProvider) VideoProcessor {
	return &DefaultVideoProcessor{
		inputDir:       inputDir,
		outputDir:      outputDir,
		config:         config,
		configProvider: configProvider,
		pathGenerator:  NewDefaultPathGenerator(inputDir, outputDir),
		converter:      NewFFmpegVideoConverter(),
		clipCalculator: NewDefaultClipCalculator(),
	}
}

func NewDefaultVideoProcessor(inputDir, outputDir string) VideoProcessor {
	defaultConfig := GetDefaultConfig()
	return &DefaultVideoProcessor{
		inputDir:       inputDir,
		outputDir:      outputDir,
		config:         defaultConfig,
		configProvider: NewDefaultConfigProvider(),
		pathGenerator:  NewDefaultPathGenerator(inputDir, outputDir),
		converter:      NewFFmpegVideoConverter(),
		clipCalculator: NewDefaultClipCalculator(),
	}
}

func (dvp *DefaultVideoProcessor) isVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, format := range supportedFormats {
		if ext == format {
			return true
		}
	}
	return false
}

func (dvp *DefaultVideoProcessor) ProcessVideo(inputPath string) error {
	fmt.Printf("\n🎬 开始处理视频: %s\n", inputPath)

	configs := dvp.configProvider.GetConfigs()

	// 使用并发处理多个配置
	return dvp.processVideoWithConcurrency(inputPath, configs)
}

func (dvp *DefaultVideoProcessor) processVideoWithConcurrency(inputPath string, configs []VideoConfig) error {
	ctx := context.Background()
	g, ctx := NewLimitedGroup(ctx, dvp.config.MaxConcurrentConfigs)

	// 用于收集处理结果
	type configResult struct {
		index int
		error error
	}
	results := make([]configResult, len(configs))
	var mu sync.Mutex

	// 并发处理每个配置
	for i, config := range configs {
		i, config := i, config // 捕获循环变量
		g.Go(func() error {
			fmt.Printf("  [%d/%d] 生成 %s 版本...\n", i+1, len(configs), config.OutputFolder)

			var err error
			if validateErr := dvp.configProvider.ValidateConfig(config); validateErr != nil {
				err = fmt.Errorf("配置验证失败 %s: %v", config.OutputFolder, validateErr)
			} else {
				err = dvp.processVideoWithConfig(inputPath, config)
			}

			// 线程安全地记录结果
			mu.Lock()
			results[i] = configResult{index: i, error: err}
			mu.Unlock()

			if err != nil {
				fmt.Printf("  ❌ 生成 %s 版本失败: %v\n", config.OutputFolder, err)
			}

			return nil // 不中断其他配置的处理
		})
	}

	// 等待所有配置处理完成
	if err := g.Wait(); err != nil {
		return fmt.Errorf("并发处理失败: %v", err)
	}

	// 检查是否有配置处理失败
	var hasErrors bool
	for _, result := range results {
		if result.error != nil {
			hasErrors = true
		}
	}

	if hasErrors {
		fmt.Printf("⚠️ 视频 %s 部分配置处理失败\n", inputPath)
	} else {
		fmt.Printf("✅ 视频 %s 处理完成\n", inputPath)
	}

	return nil // 即使部分失败也继续处理其他视频
}

func (dvp *DefaultVideoProcessor) processVideoWithConfig(inputPath string, config VideoConfig) error {
	// 获取视频总时长
	totalDuration, err := dvp.converter.GetVideoDuration(inputPath)
	if err != nil {
		return fmt.Errorf("获取视频时长失败: %v", err)
	}

	// 计算剪辑时间
	startTime, endTime, err := dvp.clipCalculator.CalculateClipTimes(totalDuration, config)
	if err != nil {
		return fmt.Errorf("计算剪辑时间失败: %v", err)
	}

	// 生成输出路径
	outputPath, err := dvp.pathGenerator.GenerateOutputPath(inputPath, config)
	if err != nil {
		return fmt.Errorf("生成输出路径失败: %v", err)
	}

	// 确保输出目录存在
	if err := dvp.pathGenerator.EnsureDir(filepath.Dir(outputPath)); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	clipDuration := endTime - startTime
	// 计算最终输出时长（应用速度后）
	finalDuration := clipDuration / config.Speed
	fmt.Printf("处理视频 [%s]: %s -> %s (截取%.2fs-%.2fs，持续%.2fs，%.1fx速，最终%.2fs，%dx%d)\n",
		config.OutputFolder, inputPath, outputPath, startTime, endTime, clipDuration, config.Speed, finalDuration, config.Width, config.Height)

	// 验证输出时长是否接近10秒
	if finalDuration < 9.5 || finalDuration > 10.5 {
		fmt.Printf("⚠️ 警告：[%s] 最终输出时长%.2fs，不是预期的10秒\n", config.OutputFolder, finalDuration)
	}

	// 创建临时配置，包含剪辑时间信息
	tempConfig := config
	tempConfig.ClipDuration = int(clipDuration)

	// 先进行剪辑
	if err := dvp.clipVideoSegment(inputPath, outputPath, startTime, clipDuration, tempConfig); err != nil {
		return fmt.Errorf("剪辑视频失败: %v", err)
	}

	fmt.Printf("✅ 成功生成 [%s]: %s (输出时长: %.2fs)\n", config.OutputFolder, outputPath, finalDuration)
	return nil
}

func (dvp *DefaultVideoProcessor) clipVideoSegment(inputPath, outputPath string, startTime, duration float64, config VideoConfig) error {
	// 分两步处理：1. 剪辑和缩放；2. 调速
	tempPath := dvp.generateTempPath(outputPath)

	// 第一步：剪辑和缩放
	if err := dvp.clipAndScale(inputPath, tempPath, startTime, duration, config); err != nil {
		return err
	}

	// 第二步：调速处理
	if err := dvp.applySpeed(tempPath, outputPath, config); err != nil {
		_ = os.Remove(tempPath) // 清理临时文件
		return err
	}

	// 清理临时文件
	_ = os.Remove(tempPath)

	// 验证最终输出时长
	if err := dvp.verifyOutputDuration(outputPath, config); err != nil {
		fmt.Printf("⚠️ 输出验证警告: %v\n", err)
	}

	return nil
}

func (dvp *DefaultVideoProcessor) generateTempPath(outputPath string) string {
	return strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + "_temp" + filepath.Ext(outputPath)
}

func (dvp *DefaultVideoProcessor) clipAndScale(inputPath, tempPath string, startTime, duration float64, config VideoConfig) error {
	args := []string{"-i", inputPath}

	// 添加时间参数
	if startTime > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.2f", startTime))
	}
	if duration > 0 {
		args = append(args, "-t", fmt.Sprintf("%.2f", duration))
	}

	// 生成视频滤镜
	videoFilter := dvp.generateVideoFilter(config)

	// Adobe Media Encoder风格的VBR编码参数
	args = append(args,
		"-vf", videoFilter,
		"-c:v", "libx264",
		"-preset", "medium",
		"-profile:v", "high",
		"-level", "4.1",
		"-b:v", fmt.Sprintf("%dk", config.VideoBitrate), // 目标比特率4Mbps
		"-maxrate", fmt.Sprintf("%dk", int(float64(config.VideoBitrate)*1.2)), // 最大比特率20%缓冲
		"-bufsize", fmt.Sprintf("%dk", config.VideoBitrate*2), // 缓冲区大小
		"-g", "50", // GOP大小
		"-keyint_min", "25", // 最小关键帧间隔
		"-sc_threshold", "40", // 场景切换阈值
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", dvp.config.AudioBitrate,
		"-ar", "48000", // 音频采样率
		"-y", tempPath,
	)

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("第一步剪辑失败: %v\n输出: %s", err, string(output))
	}
	return nil
}

// verifyOutputDuration 验证输出视频时长
func (dvp *DefaultVideoProcessor) verifyOutputDuration(outputPath string, config VideoConfig) error {
	actualDuration, err := dvp.converter.GetVideoDuration(outputPath)
	if err != nil {
		return fmt.Errorf("无法获取输出视频时长: %v", err)
	}

	expectedDuration := 10.0 // 预期10秒
	tolerance := 0.5         // 允许0.5秒误差

	if actualDuration < expectedDuration-tolerance || actualDuration > expectedDuration+tolerance {
		return fmt.Errorf("输出时长%.2fs，预期%.2fs (±%.1fs)", actualDuration, expectedDuration, tolerance)
	}

	fmt.Printf("✅ 输出时长验证通过: %.2fs\n", actualDuration)
	return nil
}

func (dvp *DefaultVideoProcessor) generateVideoFilter(config VideoConfig) string {
	if config.Width == config.Height {
		// 正方形：缩放并裁剪
		return fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d",
			config.Width, config.Height, config.Width, config.Height)
	}
	// 矩形：直接缩放
	return fmt.Sprintf("scale=%d:%d", config.Width, config.Height)
}

func (dvp *DefaultVideoProcessor) applySpeed(tempPath, outputPath string, config VideoConfig) error {
	speedPTS := 1.0 / config.Speed
	audioFilter := fmt.Sprintf("atempo=%.1f", config.Speed)

	// Adobe Media Encoder风格的高质量VBR编码
	args := []string{"-i", tempPath,
		"-filter:v", fmt.Sprintf("setpts=%.3f*PTS", speedPTS),
		"-filter:a", audioFilter,
		"-c:v", "libx264",
		"-preset", "slow", // 高质量预设
		"-profile:v", "high",
		"-level", "4.1",
		"-b:v", fmt.Sprintf("%dk", config.VideoBitrate), // VBR目标比特率4Mbps
		"-maxrate", fmt.Sprintf("%dk", int(float64(config.VideoBitrate)*1.25)), // 最大比特率25%缓冲
		"-bufsize", fmt.Sprintf("%dk", config.VideoBitrate*2), // 缓冲区大小
		"-g", "50", // GOP大小
		"-keyint_min", "25", // 最小关键帧间隔
		"-sc_threshold", "40", // 场景切换阈值
		"-bf", "3", // B帧数量
		"-b_strategy", "2", // B帧策略
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", dvp.config.AudioBitrate,
		"-ar", "48000", // 音频采样率48kHz
		"-movflags", "+faststart", // 优化网络播放
		"-y", outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("第二步调速失败: %v\n输出: %s", err, string(output))
	}
	return nil
}

func (dvp *DefaultVideoProcessor) ProcessAllVideos() error {
	// 检查输入目录是否存在
	if _, err := os.Stat(dvp.inputDir); os.IsNotExist(err) {
		return fmt.Errorf("输入目录不存在: %s", dvp.inputDir)
	}

	// 确保输出目录存在
	if err := dvp.pathGenerator.EnsureDir(dvp.outputDir); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 采用流式处理，避免一次性加载所有视频到内存
	return dvp.processVideosInBatches()
}

// processVideosInBatches 分批处理视频，控制内存占用
func (dvp *DefaultVideoProcessor) processVideosInBatches() error {
	var totalCount, successCount, errorCount int

	fmt.Printf("📁 开始扫描并分批处理视频文件...\n")

	// 分批处理视频文件
	batch := make([]string, 0, dvp.config.BatchSize)
	batchNum := 1

	err := filepath.Walk(dvp.inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查是否为视频文件
		if !dvp.isVideoFile(info.Name()) {
			fmt.Printf("跳过非视频文件: %s\n", path)
			return nil
		}

		totalCount++
		batch = append(batch, path)

		// 当批次满了或者这是最后一个文件时，处理当前批次
		if len(batch) >= dvp.config.BatchSize {
			fmt.Printf("\n🔄 处理第 %d 批 (%d 个视频)...\n", batchNum, len(batch))

			success, failed := dvp.processBatch(batch)
			successCount += success
			errorCount += failed

			fmt.Printf("✅ 第 %d 批完成：成功 %d 个，失败 %d 个\n", batchNum, success, failed)

			// 清空批次，准备下一批
			batch = batch[:0]
			batchNum++

			// 强制垃圾回收，释放内存
			runtime.GC()
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("扫描视频文件失败: %v", err)
	}

	// 处理剩余的文件
	if len(batch) > 0 {
		fmt.Printf("\n🔄 处理最后一批 (%d 个视频)...\n", len(batch))
		success, failed := dvp.processBatch(batch)
		successCount += success
		errorCount += failed
		fmt.Printf("✅ 最后一批完成：成功 %d 个，失败 %d 个\n", success, failed)
	}

	if totalCount == 0 {
		fmt.Printf("⚠️ 在 %s 目录中未找到视频文件\n", dvp.inputDir)
		return nil
	}

	fmt.Printf("\n📊 总体统计: 成功 %d 个，失败 %d 个，总计 %d 个视频\n",
		successCount, errorCount, totalCount)

	return nil
}

// processBatch 处理一批视频
func (dvp *DefaultVideoProcessor) processBatch(videoFiles []string) (int, int) {
	ctx := context.Background()
	g, ctx := NewLimitedGroup(ctx, dvp.config.MaxConcurrentVideos)

	// 用于收集处理结果
	type videoResult struct {
		path  string
		error error
	}
	results := make([]videoResult, len(videoFiles))
	var mu sync.Mutex

	// 并发处理每个视频
	for i, videoPath := range videoFiles {
		i, videoPath := i, videoPath // 捕获循环变量
		g.Go(func() error {
			err := dvp.ProcessVideo(videoPath)

			// 线程安全地记录结果
			mu.Lock()
			results[i] = videoResult{path: videoPath, error: err}
			mu.Unlock()

			return nil // 不中断其他视频的处理
		})
	}

	// 等待所有视频处理完成
	if err := g.Wait(); err != nil {
		fmt.Printf("❌ 批次处理过程中发生错误: %v\n", err)
	}

	// 统计处理结果
	var successCount, errorCount int
	for _, result := range results {
		if result.error != nil {
			errorCount++
			fmt.Printf("❌ 视频 %s 处理失败: %v\n", result.path, result.error)
		} else {
			successCount++
		}
	}

	return successCount, errorCount
}

// ==================== 工具函数 ====================

// checkFFmpeg 检查ffmpeg是否已安装
func checkFFmpeg() error {
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg未安装或不在PATH中，请先安装ffmpeg")
	}
	return nil
}

// ==================== 主函数 ====================

func main() {
	// 检查ffmpeg
	if err := checkFFmpeg(); err != nil {
		log.Fatal(err)
	}

	// 解析命令行参数
	inputDir, outputDir, configPath := parseArgs()

	// 创建配置提供者
	var configProvider ConfigProvider
	var err error
	var actualInputDir, actualOutputDir string
	var config *Config

	if configPath != "" {
		configProvider, err = NewConfigurableConfigProvider(configPath)
		if err != nil {
			log.Fatalf("加载配置文件失败: %v", err)
		}

		// 从配置文件获取目录设置
		config, _ = LoadConfig(configPath)

		// 直接使用配置文件中的设置
		actualInputDir = config.InputDir
		actualOutputDir = config.OutputDir

		fmt.Printf("📄 使用配置文件: %s\n", configPath)
	} else {
		configProvider = NewDefaultConfigProvider()
		config = GetDefaultConfig()
		actualInputDir = inputDir
		actualOutputDir = outputDir
		fmt.Printf("📄 使用默认配置\n")
	}

	// 显示配置信息
	displayConfig(actualInputDir, actualOutputDir, configProvider)

	// 创建处理器并处理视频
	processor := NewVideoProcessor(actualInputDir, actualOutputDir, config, configProvider)
	if err := processor.ProcessAllVideos(); err != nil {
		log.Fatalf("处理视频失败: %v", err)
	}

	fmt.Println("\n🎉 所有视频处理完成!")
}

func parseArgs() (string, string, string) {
	inputDir := defaultInputDir
	outputDir := defaultOutputDir
	configPath := ""

	// 智能参数解析
	if len(os.Args) == 2 && strings.HasSuffix(os.Args[1], ".json") {
		// 只有一个参数且是JSON文件，当作配置文件
		configPath = os.Args[1]
	} else {
		// 标准三参数模式：inputDir outputDir configPath
		if len(os.Args) > 1 && os.Args[1] != "" {
			inputDir = os.Args[1]
		}
		if len(os.Args) > 2 && os.Args[2] != "" {
			outputDir = os.Args[2]
		}
		if len(os.Args) > 3 && os.Args[3] != "" {
			configPath = os.Args[3]
		}
	}

	return inputDir, outputDir, configPath
}

func displayConfig(inputDir, outputDir string, configProvider ConfigProvider) {
	configs := configProvider.GetConfigs()

	fmt.Printf("🎬 视频批处理器启动 (面向对象版)\n")
	fmt.Printf("==============================\n")
	fmt.Printf("输入目录: %s\n", inputDir)
	fmt.Printf("输出目录: %s\n", outputDir)
	fmt.Printf("支持格式: %v\n", supportedFormats)
	fmt.Printf("\n📋 处理配置:\n")
	for i, config := range configs {
		fmt.Printf("  %d. %s: %dx%d, %.1fx速, %s策略, %dk比特率\n",
			i+1, config.OutputFolder, config.Width, config.Height, config.Speed, config.ClipStrategy, config.VideoBitrate)
	}
	fmt.Printf("==============================\n")
}
