package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"ai-video-clipper/internal/domain/entities"
	"ai-video-clipper/internal/domain/repositories"
)

// VideoRepositoryImpl FFmpeg视频仓储实现
type VideoRepositoryImpl struct {
	globalConfig *repositories.GlobalConfig
}

// NewVideoRepositoryImpl 创建FFmpeg视频仓储实现
func NewVideoRepositoryImpl(globalConfig *repositories.GlobalConfig) *VideoRepositoryImpl {
	return &VideoRepositoryImpl{
		globalConfig: globalConfig,
	}
}

// GetVideoDuration 获取视频时长
func (r *VideoRepositoryImpl) GetVideoDuration(ctx context.Context, videoPath string) (float64, error) {
	cmd := exec.CommandContext(ctx, "ffprobe", "-v", "quiet", "-show_entries", "format=duration", "-of", "csv=p=0", videoPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("获取视频时长失败: %v", err)
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("解析视频时长失败: %v", err)
	}

	return duration, nil
}

// ProcessVideo 处理视频
func (r *VideoRepositoryImpl) ProcessVideo(ctx context.Context, video *entities.Video, config *entities.ProcessingConfig, outputPath string) error {
	// 生成临时文件路径
	tempPath := r.generateTempPath(outputPath)

	// 计算剪辑时间
	startTime, endTime, err := config.CalculateClipTimes(video.Duration())
	if err != nil {
		return fmt.Errorf("计算剪辑时间失败: %v", err)
	}

	duration := endTime - startTime

	// 第一步：剪辑和缩放
	if err := r.clipAndScale(ctx, video.Path(), tempPath, startTime, duration, config); err != nil {
		return fmt.Errorf("剪辑和缩放失败: %v", err)
	}

	// 第二步：应用速度效果
	if err := r.applySpeed(ctx, tempPath, outputPath, config); err != nil {
		// 清理临时文件
		os.Remove(tempPath)
		return fmt.Errorf("应用速度效果失败: %v", err)
	}

	// 清理临时文件
	os.Remove(tempPath)

	return nil
}

// ListVideoFiles 列出目录中的视频文件
func (r *VideoRepositoryImpl) ListVideoFiles(ctx context.Context, inputDir string) ([]string, error) {
	var videoFiles []string
	supportedFormats := []string{".mp4", ".avi", ".mov", ".mkv", ".flv", ".wmv", ".m4v", ".3gp", ".webm"}

	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		for _, format := range supportedFormats {
			if ext == format {
				videoFiles = append(videoFiles, path)
				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历目录失败: %v", err)
	}

	return videoFiles, nil
}

// EnsureOutputDir 确保输出目录存在
func (r *VideoRepositoryImpl) EnsureOutputDir(ctx context.Context, outputDir string) error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return os.MkdirAll(outputDir, 0755)
	}
	return nil
}

// generateTempPath 生成临时文件路径
func (r *VideoRepositoryImpl) generateTempPath(outputPath string) string {
	dir := filepath.Dir(outputPath)
	filename := filepath.Base(outputPath)
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	return filepath.Join(dir, fmt.Sprintf("%s_temp%s", name, ext))
}

// clipAndScale 剪辑和缩放视频
func (r *VideoRepositoryImpl) clipAndScale(ctx context.Context, inputPath, outputPath string, startTime, duration float64, config *entities.ProcessingConfig) error {
	// 生成视频滤镜
	videoFilter := r.generateVideoFilter(config)

	// 获取质量参数
	qualityParams := r.getVideoQualityParams(r.getOptimalQualityPreset(), config.VideoBitrate())
	qualityArgs := r.buildVideoQualityArgs(qualityParams, config.VideoBitrate(), true)

	// 构建FFmpeg命令
	args := []string{
		"-i", inputPath,
		"-ss", fmt.Sprintf("%.2f", startTime),
		"-t", fmt.Sprintf("%.2f", duration),
		"-vf", videoFilter,
		"-c:v", "libx264",
	}

	// 添加质量参数
	args = append(args, qualityArgs...)

	// 添加音频参数
	args = append(args,
		"-c:a", "aac",
		"-b:a", r.globalConfig.AudioBitrate,
		"-ar", "48000",
		"-y", outputPath,
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("FFmpeg剪辑失败: %v", err)
	}

	return nil
}

// applySpeed 应用速度效果
func (r *VideoRepositoryImpl) applySpeed(ctx context.Context, inputPath, outputPath string, config *entities.ProcessingConfig) error {
	// 计算速度滤镜
	videoSpeedFilter := fmt.Sprintf("setpts=%.2f*PTS", 1.0/config.Speed())
	audioSpeedFilter := fmt.Sprintf("atempo=%.2f", config.Speed())

	// 获取质量参数
	qualityParams := r.getVideoQualityParams(r.getOptimalQualityPreset(), config.VideoBitrate())
	qualityArgs := r.buildVideoQualityArgs(qualityParams, config.VideoBitrate(), true)

	// 构建FFmpeg命令
	args := []string{
		"-i", inputPath,
		"-vf", videoSpeedFilter,
		"-af", audioSpeedFilter,
		"-c:v", "libx264",
	}

	// 添加质量参数
	args = append(args, qualityArgs...)

	// 添加音频参数
	args = append(args,
		"-c:a", "aac",
		"-b:a", r.globalConfig.AudioBitrate,
		"-ar", "48000",
		"-y", outputPath,
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("FFmpeg速度调整失败: %v", err)
	}

	return nil
}

// generateVideoFilter 生成视频滤镜
func (r *VideoRepositoryImpl) generateVideoFilter(config *entities.ProcessingConfig) string {
	return fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:black",
		config.Width(), config.Height(), config.Width(), config.Height())
}

// VideoQualityPreset 视频质量预设类型
type VideoQualityPreset string

const (
	QualityFast  VideoQualityPreset = "fast"
	QualityHigh  VideoQualityPreset = "high"
	QualityUltra VideoQualityPreset = "ultra"
)

// VideoQualityParams 视频质量参数
type VideoQualityParams struct {
	Preset      string
	CRF         int
	Profile     string
	Level       string
	BFrames     int
	RefFrames   int
	MEMethod    string
	Subme       int
	TrellisME   int
	Keyint      int
	KeyintMin   int
	ScThreshold int
	UsePsy      bool
	PsyRD       string
	PsyTrellis  string
	Deblock     string
	NoMbtree    bool
	UseCABAC    bool
	Mixed8x8dct bool
}

// getVideoQualityParams 获取视频质量参数
func (r *VideoRepositoryImpl) getVideoQualityParams(preset VideoQualityPreset, bitrate int) VideoQualityParams {
	switch preset {
	case QualityFast:
		return VideoQualityParams{
			Preset:      "fast",
			CRF:         23,
			Profile:     "high",
			Level:       "4.1",
			BFrames:     3,
			RefFrames:   3,
			MEMethod:    "hex",
			Subme:       6,
			TrellisME:   1,
			Keyint:      250,
			KeyintMin:   25,
			ScThreshold: 40,
			UsePsy:      true,
			PsyRD:       "1.0:0.0",
			PsyTrellis:  "0.0",
			Deblock:     "1:0:0",
			NoMbtree:    false,
			UseCABAC:    true,
			Mixed8x8dct: true,
		}
	case QualityHigh:
		return VideoQualityParams{
			Preset:      "slow",
			CRF:         20,
			Profile:     "high",
			Level:       "4.1",
			BFrames:     8,
			RefFrames:   5,
			MEMethod:    "umh",
			Subme:       8,
			TrellisME:   2,
			Keyint:      250,
			KeyintMin:   25,
			ScThreshold: 40,
			UsePsy:      true,
			PsyRD:       "1.0:0.1",
			PsyTrellis:  "0.2",
			Deblock:     "1:0:0",
			NoMbtree:    false,
			UseCABAC:    true,
			Mixed8x8dct: true,
		}
	case QualityUltra:
		return VideoQualityParams{
			Preset:      "veryslow",
			CRF:         18,
			Profile:     "high",
			Level:       "4.1",
			BFrames:     16,
			RefFrames:   8,
			MEMethod:    "tesa",
			Subme:       10,
			TrellisME:   2,
			Keyint:      250,
			KeyintMin:   25,
			ScThreshold: 40,
			UsePsy:      true,
			PsyRD:       "1.0:0.15",
			PsyTrellis:  "0.25",
			Deblock:     "1:0:0",
			NoMbtree:    false,
			UseCABAC:    true,
			Mixed8x8dct: true,
		}
	default:
		return r.getVideoQualityParams(QualityHigh, bitrate)
	}
}

// buildVideoQualityArgs 构建视频质量参数
func (r *VideoRepositoryImpl) buildVideoQualityArgs(params VideoQualityParams, bitrate int, useCRF bool) []string {
	args := []string{
		"-preset", params.Preset,
		"-profile:v", params.Profile,
		"-level", params.Level,
		"-pix_fmt", "yuv420p",
	}

	if useCRF {
		args = append(args, "-crf", strconv.Itoa(params.CRF))
		args = append(args, "-maxrate", fmt.Sprintf("%dk", int(float64(bitrate)*1.5)))
		args = append(args, "-bufsize", fmt.Sprintf("%dk", bitrate*2))
	} else {
		args = append(args, "-b:v", fmt.Sprintf("%dk", bitrate))
		args = append(args, "-maxrate", fmt.Sprintf("%dk", int(float64(bitrate)*1.2)))
		args = append(args, "-bufsize", fmt.Sprintf("%dk", bitrate*2))
	}

	return args
}

// getOptimalQualityPreset 获取最优质量预设
func (r *VideoRepositoryImpl) getOptimalQualityPreset() VideoQualityPreset {
	switch r.globalConfig.QualityPreset {
	case "fast":
		return QualityFast
	case "high":
		return QualityHigh
	case "ultra":
		return QualityUltra
	default:
		return QualityHigh
	}
}
