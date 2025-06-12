package repositories

import (
	"ai-video-clipper/internal/domain/entities"
	"context"
)

// VideoRepository 视频仓储接口
type VideoRepository interface {
	// GetVideoDuration 获取视频时长
	GetVideoDuration(ctx context.Context, videoPath string) (float64, error)

	// ProcessVideo 处理视频
	ProcessVideo(ctx context.Context, video *entities.Video, config *entities.ProcessingConfig, outputPath string) error

	// ListVideoFiles 列出目录中的视频文件
	ListVideoFiles(ctx context.Context, inputDir string) ([]string, error)

	// EnsureOutputDir 确保输出目录存在
	EnsureOutputDir(ctx context.Context, outputDir string) error
}
