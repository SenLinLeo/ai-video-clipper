package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config 全局配置
type Config struct {
	InputDir             string        `json:"inputDir"`
	OutputDir            string        `json:"outputDir"`
	AudioBitrate         string        `json:"audioBitrate"`
	MaxConcurrentVideos  int           `json:"maxConcurrentVideos"`
	MaxConcurrentConfigs int           `json:"maxConcurrentConfigs"`
	BatchSize            int           `json:"batchSize"`
	VideoConfigs         []VideoConfig `json:"videoConfigs"`
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	// 如果配置文件不存在，返回默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return GetDefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, configPath string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *Config {
	return &Config{
		InputDir:             "/Volumes/Data/youtube-download",
		OutputDir:            "/Volumes/Data/youtube-download/output",
		AudioBitrate:         "112k",
		MaxConcurrentVideos:  10,
		MaxConcurrentConfigs: 50,
		BatchSize:            20,
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
