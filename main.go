package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// 默认输入和输出目录
	defaultInputDir  = "input"
	defaultOutputDir = "output"
	// 音频比特率
	audioBitrate = "112k"
)

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
	ConvertVideo(inputPath, outputPath string, config VideoConfig) error
	GetVideoDuration(videoPath string) (float64, error)
}

// ClipCalculator 剪辑计算器接口
type ClipCalculator interface {
	CalculateClipTimes(totalDuration float64, config VideoConfig) (float64, float64, error)
}

// ==================== 数据结构 ====================

// VideoConfig 视频配置
type VideoConfig struct {
	Width          int     // 视频宽度
	Height         int     // 视频高度
	ClipDuration   int     // 剪辑时长（秒）
	Speed          float64 // 播放速度倍数
	VideoBitrate   int     // 视频比特率(kbps)
	ClipStrategy   string  // 剪辑策略：last_segments 或 middle_segments
	OutputSuffix   string  // 输出文件后缀
	OutputFolder   string  // 输出文件夹
}

// ==================== 接口实现 ====================

// DefaultConfigProvider 默认配置提供者
type DefaultConfigProvider struct{}

func NewDefaultConfigProvider() ConfigProvider {
	return &DefaultConfigProvider{}
}

func (dcp *DefaultConfigProvider) GetConfigs() []VideoConfig {
	return []VideoConfig{
		{
			Width:        1008,
			Height:       1008,
			ClipDuration: 25,
			Speed:        2.5,
			VideoBitrate: 1000,
			ClipStrategy: "last_segments",
			OutputSuffix: "_square_last",
			OutputFolder: "1008x1008_2.5x",
		},
		{
			Width:        1008,
			Height:       762,
			ClipDuration: 25,
			Speed:        2.5,
			VideoBitrate: 1000,
			ClipStrategy: "last_segments",
			OutputSuffix: "_rect_last",
			OutputFolder: "1008x762_2.5x",
		},
		{
			Width:        1008,
			Height:       1008,
			ClipDuration: 20,
			Speed:        2.0,
			VideoBitrate: 1000,
			ClipStrategy: "middle_segments",
			OutputSuffix: "_square_middle",
			OutputFolder: "1008x1008_2.0x",
		},
		{
			Width:        1008,
			Height:       762,
			ClipDuration: 20,
			Speed:        2.0,
			VideoBitrate: 1000,
			ClipStrategy: "middle_segments",
			OutputSuffix: "_rect_middle",
			OutputFolder: "1008x762_2.0x",
		},
	}
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
	case "last_segments":
		// 截取倒数第5秒的前N秒
		endTime = totalDuration - 5
		startTime = endTime - float64(config.ClipDuration)
		if startTime < 0 {
			startTime = 0
		}
	case "middle_segments":
		// 截取开始前5秒至结束前5秒之间的N秒内容
		availableStart := 5.0 // 开始前5秒后
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

func (fvc *FFmpegVideoConverter) ConvertVideo(inputPath, outputPath string, config VideoConfig) error {
	// 第一步：剪辑和基础转换
	tempPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + "_temp" + filepath.Ext(outputPath)
	if err := fvc.clipAndConvert(inputPath, tempPath, config); err != nil {
		return fmt.Errorf("剪辑转换失败: %v", err)
	}
	
	// 第二步：应用速度效果和最终压缩
	if err := fvc.applySpeedAndCompress(tempPath, outputPath, config); err != nil {
		os.Remove(tempPath) // 清理临时文件
		return fmt.Errorf("应用速度效果失败: %v", err)
	}
	
	// 清理临时文件
	os.Remove(tempPath)
	
	return nil
}

func (fvc *FFmpegVideoConverter) clipAndConvert(inputPath, outputPath string, config VideoConfig) error {
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
	
	args = append(args,
		"-vf", videoFilter,
		"-c:v", "libx264",
		"-preset", "medium",
		"-b:v", fmt.Sprintf("%dk", config.VideoBitrate),
		"-maxrate", fmt.Sprintf("%dk", config.VideoBitrate),
		"-bufsize", "2M",
		"-c:a", "aac",
		"-b:a", audioBitrate,
		"-y", outputPath,
	)
	
	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg剪辑失败: %v\n输出: %s", err, string(output))
	}
	
	return nil
}

func (fvc *FFmpegVideoConverter) applySpeedAndCompress(inputPath, outputPath string, config VideoConfig) error {
	speedPTS := 1.0 / config.Speed // 视频PTS调整
	
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-filter:v", fmt.Sprintf("setpts=%.3f*PTS", speedPTS),
		"-filter:a", fmt.Sprintf("atempo=%.1f", config.Speed),
		"-c:v", "libx264",
		"-preset", "slow",      // 高质量压缩
		"-crf", "20",           // 质量因子
		"-profile:v", "high",
		"-level", "4.1",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", audioBitrate,
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
	inputDir      string
	outputDir     string
	configProvider ConfigProvider
	pathGenerator PathGenerator
	converter     VideoConverter
	clipCalculator ClipCalculator
}

func NewDefaultVideoProcessor(inputDir, outputDir string) VideoProcessor {
	return &DefaultVideoProcessor{
		inputDir:       inputDir,
		outputDir:      outputDir,
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
	
	// 为每个视频生成四个版本
	for i, config := range configs {
		fmt.Printf("  [%d/%d] 生成 %s 版本...\n", i+1, len(configs), config.OutputFolder)
		
		if err := dvp.configProvider.ValidateConfig(config); err != nil {
			fmt.Printf("  ❌ 配置验证失败 %s: %v\n", config.OutputFolder, err)
			continue
		}
		
		if err := dvp.processVideoWithConfig(inputPath, config); err != nil {
			fmt.Printf("  ❌ 生成 %s 版本失败: %v\n", config.OutputFolder, err)
			// 继续处理其他版本
		}
	}
	
	fmt.Printf("✅ 视频 %s 处理完成\n", inputPath)
	return nil
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
	fmt.Printf("处理视频 [%s]: %s -> %s (截取%.2fs-%.2fs，持续%.2fs，%.1fx速，%dx%d)\n", 
		config.OutputFolder, inputPath, outputPath, startTime, endTime, clipDuration, config.Speed, config.Width, config.Height)
	
	// 创建临时配置，包含剪辑时间信息
	tempConfig := config
	tempConfig.ClipDuration = int(clipDuration)
	
	// 先进行剪辑
	if err := dvp.clipVideoSegment(inputPath, outputPath, startTime, clipDuration, tempConfig); err != nil {
		return fmt.Errorf("剪辑视频失败: %v", err)
	}
	
	fmt.Printf("✅ 成功生成 [%s]: %s\n", config.OutputFolder, outputPath)
	return nil
}

func (dvp *DefaultVideoProcessor) clipVideoSegment(inputPath, outputPath string, startTime, duration float64, config VideoConfig) error {
	// 分两步处理：1. 剪辑和缩放；2. 调速
	
	// 第一步：剪辑和缩放
	tempPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + "_temp" + filepath.Ext(outputPath)
	
	args1 := []string{"-i", inputPath}
	
	// 添加时间参数
	if startTime > 0 {
		args1 = append(args1, "-ss", fmt.Sprintf("%.2f", startTime))
	}
	if duration > 0 {
		args1 = append(args1, "-t", fmt.Sprintf("%.2f", duration))
	}
	
	// 生成视频滤镜（只缩放，不调速）
	var videoFilter string
	if config.Width == config.Height {
		// 正方形：缩放并裁剪
		videoFilter = fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d", 
			config.Width, config.Height, config.Width, config.Height)
	} else {
		// 矩形：直接缩放
		videoFilter = fmt.Sprintf("scale=%d:%d", config.Width, config.Height)
	}
	
	args1 = append(args1,
		"-vf", videoFilter,
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", audioBitrate,
		"-y", tempPath,
	)
	
	cmd1 := exec.Command("ffmpeg", args1...)
	output1, err := cmd1.CombinedOutput()
	if err != nil {
		return fmt.Errorf("第一步剪辑失败: %v\n输出: %s", err, string(output1))
	}
	
	// 第二步：调速处理
	speedPTS := 1.0 / config.Speed
	
	// 处理音频速度调整
	var audioFilter string
	if config.Speed <= 2.0 {
		audioFilter = fmt.Sprintf("atempo=%.1f", config.Speed)
	} else {
		// 对于大于2.0的速度，分级处理
		audioFilter = "atempo=2.0,atempo=" + fmt.Sprintf("%.2f", config.Speed/2.0)
	}
	
	args2 := []string{"-i", tempPath,
		"-filter:v", fmt.Sprintf("setpts=%.3f*PTS", speedPTS),
		"-filter:a", audioFilter,
		"-c:v", "libx264",
		"-preset", "slow",
		"-crf", "20",
		"-profile:v", "high",
		"-level", "4.1",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", audioBitrate,
		"-movflags", "+faststart",
		"-y", outputPath,
	}
	
	cmd2 := exec.Command("ffmpeg", args2...)
	output2, err := cmd2.CombinedOutput()
	if err != nil {
		os.Remove(tempPath) // 清理临时文件
		return fmt.Errorf("第二步调速失败: %v\n输出: %s", err, string(output2))
	}
	
	// 清理临时文件
	os.Remove(tempPath)
	
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
	
	// 遍历输入目录
	return filepath.Walk(dvp.inputDir, func(path string, info os.FileInfo, err error) error {
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
		
		// 处理视频
		return dvp.ProcessVideo(path)
	})
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
	inputDir := defaultInputDir
	outputDir := defaultOutputDir
	
	if len(os.Args) > 1 {
		inputDir = os.Args[1]
	}
	if len(os.Args) > 2 {
		outputDir = os.Args[2]
	}
	
	// 创建配置提供者来显示配置信息
	configProvider := NewDefaultConfigProvider()
	configs := configProvider.GetConfigs()
	
	fmt.Printf("🎬 视频批处理器启动 (面向对象版)\n")
	fmt.Printf("==============================\n")
	fmt.Printf("输入目录: %s\n", inputDir)
	fmt.Printf("输出目录: %s\n", outputDir)
	fmt.Printf("音频比特率: %s\n", audioBitrate)
	fmt.Printf("支持格式: %v\n", supportedFormats)
	fmt.Printf("\n📋 处理配置:\n")
	for i, config := range configs {
		fmt.Printf("  %d. %s: %dx%d, %.1fx速, %s策略\n", 
			i+1, config.OutputFolder, config.Width, config.Height, config.Speed, config.ClipStrategy)
	}
	fmt.Printf("==============================\n")
	
	// 创建处理器并处理视频
	processor := NewDefaultVideoProcessor(inputDir, outputDir)
	if err := processor.ProcessAllVideos(); err != nil {
		log.Fatalf("处理视频失败: %v", err)
	}
	
	fmt.Println("\n🎉 所有视频处理完成!")
} 