package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"ai-video-clipper/internal/domain/entities"
	"ai-video-clipper/internal/domain/repositories"
)

// ConfigRepositoryImpl 配置仓储实现
type ConfigRepositoryImpl struct {
	configPath string
}

// NewConfigRepositoryImpl 创建配置仓储实现
func NewConfigRepositoryImpl(configPath string) *ConfigRepositoryImpl {
	return &ConfigRepositoryImpl{
		configPath: configPath,
	}
}

// LoadProcessingConfigs 加载处理配置列表
func (r *ConfigRepositoryImpl) LoadProcessingConfigs(ctx context.Context) ([]*entities.ProcessingConfig, error) {
	config, err := r.loadConfig()
	if err != nil {
		return nil, err
	}

	var processingConfigs []*entities.ProcessingConfig
	for _, vc := range config.VideoConfigs {
		pc, err := entities.NewProcessingConfig(
			vc.Width,
			vc.Height,
			vc.ClipDuration,
			vc.Speed,
			vc.VideoBitrate,
			vc.ClipStrategy,
			vc.OutputSuffix,
			vc.OutputFolder,
		)
		if err != nil {
			return nil, fmt.Errorf("创建处理配置失败: %v", err)
		}
		processingConfigs = append(processingConfigs, pc)
	}

	return processingConfigs, nil
}

// GetGlobalConfig 获取全局配置
func (r *ConfigRepositoryImpl) GetGlobalConfig(ctx context.Context) (*repositories.GlobalConfig, error) {
	config, err := r.loadConfig()
	if err != nil {
		return nil, err
	}

	return &repositories.GlobalConfig{
		InputDir:             config.InputDir,
		OutputDir:            config.OutputDir,
		AudioBitrate:         config.AudioBitrate,
		MaxConcurrentVideos:  config.MaxConcurrentVideos,
		MaxConcurrentConfigs: config.MaxConcurrentConfigs,
		BatchSize:            config.BatchSize,
		QualityPreset:        config.QualityPreset,
	}, nil
}

// loadConfig 加载配置文件
func (r *ConfigRepositoryImpl) loadConfig() (*Config, error) {
	// 如果配置文件不存在，返回默认配置
	if _, err := os.Stat(r.configPath); os.IsNotExist(err) {
		return getDefaultConfig(), nil
	}

	data, err := os.ReadFile(r.configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// Config 配置结构体（复用原有结构）
type Config struct {
	InputDir             string        `json:"inputDir"`
	OutputDir            string        `json:"outputDir"`
	AudioBitrate         string        `json:"audioBitrate"`
	MaxConcurrentVideos  int           `json:"maxConcurrentVideos"`
	MaxConcurrentConfigs int           `json:"maxConcurrentConfigs"`
	BatchSize            int           `json:"batchSize"`
	QualityPreset        string        `json:"qualityPreset"`
	VideoConfigs         []VideoConfig `json:"videoConfigs"`
}

// VideoConfig 视频配置结构体（复用原有结构）
type VideoConfig struct {
	Width        int     `json:"Width"`
	Height       int     `json:"Height"`
	ClipDuration int     `json:"ClipDuration"`
	Speed        float64 `json:"Speed"`
	VideoBitrate int     `json:"VideoBitrate"`
	ClipStrategy string  `json:"ClipStrategy"`
	OutputSuffix string  `json:"OutputSuffix"`
	OutputFolder string  `json:"OutputFolder"`
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *Config {
	return &Config{
		InputDir:             "/Volumes/Data/youtube-download",
		OutputDir:            "/Volumes/Data/youtube-download/output",
		AudioBitrate:         "112k",
		MaxConcurrentVideos:  10,
		MaxConcurrentConfigs: 50,
		BatchSize:            20,
		QualityPreset:        "high",
		VideoConfigs: []VideoConfig{
			{
				Width:        1008,
				Height:       1008,
				ClipDuration: 20,
				Speed:        2.0,
				VideoBitrate: 4000,
				ClipStrategy: "start_segments",
				OutputSuffix: "_square_start",
				OutputFolder: "1008x1008_start",
			},
			{
				Width:        1008,
				Height:       762,
				ClipDuration: 20,
				Speed:        2.0,
				VideoBitrate: 4000,
				ClipStrategy: "start_segments",
				OutputSuffix: "_rect_start",
				OutputFolder: "1008x762_start",
			},
			{
				Width:        1008,
				Height:       1008,
				ClipDuration: 20,
				Speed:        2.0,
				VideoBitrate: 4000,
				ClipStrategy: "end_segments",
				OutputSuffix: "_square_end",
				OutputFolder: "1008x1008_end",
			},
			{
				Width:        1008,
				Height:       762,
				ClipDuration: 20,
				Speed:        2.0,
				VideoBitrate: 4000,
				ClipStrategy: "end_segments",
				OutputSuffix: "_rect_end",
				OutputFolder: "1008x762_end",
			},
		},
	}
}
