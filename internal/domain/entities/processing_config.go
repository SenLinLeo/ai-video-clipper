package entities

import (
	"fmt"
)

// ProcessingConfig 视频处理配置实体
type ProcessingConfig struct {
	width        int
	height       int
	clipDuration int
	speed        float64
	videoBitrate int
	clipStrategy string
	outputSuffix string
	outputFolder string
}

// NewProcessingConfig 创建新的处理配置
func NewProcessingConfig(width, height, clipDuration int, speed float64, videoBitrate int,
	clipStrategy, outputSuffix, outputFolder string) (*ProcessingConfig, error) {

	config := &ProcessingConfig{
		width:        width,
		height:       height,
		clipDuration: clipDuration,
		speed:        speed,
		videoBitrate: videoBitrate,
		clipStrategy: clipStrategy,
		outputSuffix: outputSuffix,
		outputFolder: outputFolder,
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// validate 验证配置有效性
func (pc *ProcessingConfig) validate() error {
	if pc.width <= 0 || pc.height <= 0 {
		return fmt.Errorf("分辨率无效: %dx%d", pc.width, pc.height)
	}
	if pc.speed <= 0 {
		return fmt.Errorf("播放速度无效: %.1f", pc.speed)
	}
	if pc.clipDuration <= 0 {
		return fmt.Errorf("剪辑时长无效: %d", pc.clipDuration)
	}
	if pc.videoBitrate <= 0 {
		return fmt.Errorf("视频比特率无效: %d", pc.videoBitrate)
	}
	if pc.clipStrategy != "start_segments" && pc.clipStrategy != "end_segments" &&
		pc.clipStrategy != "middle_segments" && pc.clipStrategy != "last_segments" {
		return fmt.Errorf("剪辑策略无效: %s", pc.clipStrategy)
	}
	return nil
}

// Width 获取宽度
func (pc *ProcessingConfig) Width() int {
	return pc.width
}

// Height 获取高度
func (pc *ProcessingConfig) Height() int {
	return pc.height
}

// ClipDuration 获取剪辑时长
func (pc *ProcessingConfig) ClipDuration() int {
	return pc.clipDuration
}

// Speed 获取播放速度
func (pc *ProcessingConfig) Speed() float64 {
	return pc.speed
}

// VideoBitrate 获取视频比特率
func (pc *ProcessingConfig) VideoBitrate() int {
	return pc.videoBitrate
}

// ClipStrategy 获取剪辑策略
func (pc *ProcessingConfig) ClipStrategy() string {
	return pc.clipStrategy
}

// OutputSuffix 获取输出后缀
func (pc *ProcessingConfig) OutputSuffix() string {
	return pc.outputSuffix
}

// OutputFolder 获取输出文件夹
func (pc *ProcessingConfig) OutputFolder() string {
	return pc.outputFolder
}

// CalculateClipTimes 计算剪辑时间点
func (pc *ProcessingConfig) CalculateClipTimes(totalDuration float64) (startTime, endTime float64, err error) {
	switch pc.clipStrategy {
	case "start_segments":
		startTime = 5.0
		endTime = startTime + float64(pc.clipDuration)
		if endTime > totalDuration {
			endTime = totalDuration
		}
	case "end_segments":
		endTime = totalDuration - 5
		startTime = endTime - float64(pc.clipDuration)
		if startTime < 0 {
			startTime = 0
		}
	case "middle_segments":
		middle := totalDuration / 2
		startTime = middle - float64(pc.clipDuration)/2
		endTime = middle + float64(pc.clipDuration)/2
		if startTime < 0 {
			startTime = 0
			endTime = float64(pc.clipDuration)
		}
		if endTime > totalDuration {
			endTime = totalDuration
			startTime = totalDuration - float64(pc.clipDuration)
		}
	case "last_segments":
		startTime = totalDuration - float64(pc.clipDuration)
		endTime = totalDuration
		if startTime < 0 {
			startTime = 0
		}
	default:
		return 0, 0, fmt.Errorf("不支持的剪辑策略: %s", pc.clipStrategy)
	}

	if startTime >= endTime {
		return 0, 0, fmt.Errorf("计算的剪辑时间无效: start=%.2f, end=%.2f", startTime, endTime)
	}

	return startTime, endTime, nil
}
