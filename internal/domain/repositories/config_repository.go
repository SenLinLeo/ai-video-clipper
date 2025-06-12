package repositories

import (
	"ai-video-clipper/internal/domain/entities"
	"context"
)

// ConfigRepository 配置仓储接口
type ConfigRepository interface {
	// LoadProcessingConfigs 加载处理配置列表
	LoadProcessingConfigs(ctx context.Context) ([]*entities.ProcessingConfig, error)

	// GetGlobalConfig 获取全局配置
	GetGlobalConfig(ctx context.Context) (*GlobalConfig, error)
}

// GlobalConfig 全局配置
type GlobalConfig struct {
	InputDir             string
	OutputDir            string
	AudioBitrate         string
	MaxConcurrentVideos  int
	MaxConcurrentConfigs int
	BatchSize            int
	QualityPreset        string
}
