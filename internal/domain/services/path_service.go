package services

import (
	"ai-video-clipper/internal/domain/entities"
	"fmt"
	"path/filepath"
	"strings"
)

// PathService 路径服务
type PathService struct {
	inputDir  string
	outputDir string
}

// NewPathService 创建路径服务
func NewPathService(inputDir, outputDir string) *PathService {
	return &PathService{
		inputDir:  inputDir,
		outputDir: outputDir,
	}
}

// GenerateOutputPath 生成输出路径
func (ps *PathService) GenerateOutputPath(video *entities.Video, config *entities.ProcessingConfig) (string, error) {
	relPath, err := filepath.Rel(ps.inputDir, video.Path())
	if err != nil {
		return "", fmt.Errorf("计算相对路径失败: %v", err)
	}

	dir := filepath.Dir(relPath)
	filename := filepath.Base(relPath)
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	// 生成新文件名：原名 + 后缀 + 扩展名
	newFilename := fmt.Sprintf("%s%s%s", name, config.OutputSuffix(), ext)

	// 生成完整输出路径
	var outputPath string
	if dir == "." {
		outputPath = filepath.Join(ps.outputDir, config.OutputFolder(), newFilename)
	} else {
		outputPath = filepath.Join(ps.outputDir, config.OutputFolder(), dir, newFilename)
	}

	return outputPath, nil
}

// GenerateTempPath 生成临时文件路径
func (ps *PathService) GenerateTempPath(outputPath string) string {
	dir := filepath.Dir(outputPath)
	filename := filepath.Base(outputPath)
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	return filepath.Join(dir, fmt.Sprintf("%s_temp%s", name, ext))
}
