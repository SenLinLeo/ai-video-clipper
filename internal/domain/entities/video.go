package entities

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Video 视频领域实体
type Video struct {
	path     string
	duration float64
	format   string
}

// NewVideo 创建新的视频实体
func NewVideo(path string) (*Video, error) {
	if path == "" {
		return nil, fmt.Errorf("视频路径不能为空")
	}

	ext := strings.ToLower(filepath.Ext(path))
	if !isValidVideoFormat(ext) {
		return nil, fmt.Errorf("不支持的视频格式: %s", ext)
	}

	return &Video{
		path:   path,
		format: ext,
	}, nil
}

// Path 获取视频路径
func (v *Video) Path() string {
	return v.path
}

// Duration 获取视频时长
func (v *Video) Duration() float64 {
	return v.duration
}

// SetDuration 设置视频时长
func (v *Video) SetDuration(duration float64) error {
	if duration <= 0 {
		return fmt.Errorf("视频时长必须大于0")
	}
	v.duration = duration
	return nil
}

// Format 获取视频格式
func (v *Video) Format() string {
	return v.format
}

// CanClip 检查是否可以剪辑
func (v *Video) CanClip(clipDuration float64) bool {
	return v.duration > clipDuration+10 // 至少比剪辑时长多10秒
}

// isValidVideoFormat 检查是否为支持的视频格式
func isValidVideoFormat(ext string) bool {
	supportedFormats := []string{".mp4", ".avi", ".mov", ".mkv", ".flv", ".wmv", ".m4v", ".3gp", ".webm"}
	for _, format := range supportedFormats {
		if ext == format {
			return true
		}
	}
	return false
}
