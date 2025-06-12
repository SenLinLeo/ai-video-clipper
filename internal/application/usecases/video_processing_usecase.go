package usecases

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"ai-video-clipper/internal/application/dto"
	"ai-video-clipper/internal/domain/entities"
	"ai-video-clipper/internal/domain/repositories"
	"ai-video-clipper/internal/domain/services"
)

// VideoProcessingUseCase 视频处理用例
type VideoProcessingUseCase struct {
	videoRepo   repositories.VideoRepository
	configRepo  repositories.ConfigRepository
	pathService *services.PathService
}

// NewVideoProcessingUseCase 创建视频处理用例
func NewVideoProcessingUseCase(
	videoRepo repositories.VideoRepository,
	configRepo repositories.ConfigRepository,
	pathService *services.PathService,
) *VideoProcessingUseCase {
	return &VideoProcessingUseCase{
		videoRepo:   videoRepo,
		configRepo:  configRepo,
		pathService: pathService,
	}
}

// ProcessSingleVideo 处理单个视频
func (uc *VideoProcessingUseCase) ProcessSingleVideo(ctx context.Context, req *dto.ProcessVideoRequest) (*dto.ProcessVideoResponse, error) {
	// 创建视频实体
	video, err := entities.NewVideo(req.InputPath)
	if err != nil {
		return &dto.ProcessVideoResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("创建视频实体失败: %v", err),
		}, nil
	}

	// 获取视频时长
	duration, err := uc.videoRepo.GetVideoDuration(ctx, video.Path())
	if err != nil {
		return &dto.ProcessVideoResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("获取视频时长失败: %v", err),
		}, nil
	}

	if err := video.SetDuration(duration); err != nil {
		return &dto.ProcessVideoResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("设置视频时长失败: %v", err),
		}, nil
	}

	// 加载处理配置
	configs, err := uc.configRepo.LoadProcessingConfigs(ctx)
	if err != nil {
		return &dto.ProcessVideoResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("加载配置失败: %v", err),
		}, nil
	}

	var processedFiles []string
	var errors []string

	// 处理每个配置
	for _, config := range configs {
		if !video.CanClip(float64(config.ClipDuration())) {
			errors = append(errors, fmt.Sprintf("视频时长不足，无法使用配置 %s", config.OutputSuffix()))
			continue
		}

		outputPath, err := uc.pathService.GenerateOutputPath(video, config)
		if err != nil {
			errors = append(errors, fmt.Sprintf("生成输出路径失败: %v", err))
			continue
		}

		// 确保输出目录存在
		if err := uc.videoRepo.EnsureOutputDir(ctx, filepath.Dir(outputPath)); err != nil {
			errors = append(errors, fmt.Sprintf("创建输出目录失败: %v", err))
			continue
		}

		// 处理视频
		if err := uc.videoRepo.ProcessVideo(ctx, video, config, outputPath); err != nil {
			errors = append(errors, fmt.Sprintf("处理视频失败 [%s]: %v", config.OutputSuffix(), err))
			continue
		}

		processedFiles = append(processedFiles, outputPath)
		log.Printf("成功处理视频: %s -> %s", video.Path(), outputPath)
	}

	success := len(processedFiles) > 0
	var errorMessage string
	if len(errors) > 0 {
		errorMessage = strings.Join(errors, "; ")
	}

	return &dto.ProcessVideoResponse{
		Success:        success,
		ProcessedFiles: processedFiles,
		ErrorMessage:   errorMessage,
	}, nil
}

// ProcessBatch 批量处理视频
func (uc *VideoProcessingUseCase) ProcessBatch(ctx context.Context, req *dto.BatchProcessRequest) (*dto.BatchProcessResponse, error) {
	// 获取全局配置
	globalConfig, err := uc.configRepo.GetGlobalConfig(ctx)
	if err != nil {
		return &dto.BatchProcessResponse{
			ErrorMessage: fmt.Sprintf("获取全局配置失败: %v", err),
		}, nil
	}

	// 使用配置中的目录，而不是请求中的目录
	inputDir := globalConfig.InputDir
	outputDir := globalConfig.OutputDir

	// 更新路径服务
	uc.pathService = services.NewPathService(inputDir, outputDir)

	// 列出视频文件
	videoFiles, err := uc.videoRepo.ListVideoFiles(ctx, inputDir)
	if err != nil {
		return &dto.BatchProcessResponse{
			ErrorMessage: fmt.Sprintf("列出视频文件失败: %v", err),
		}, nil
	}

	totalVideos := len(videoFiles)
	if totalVideos == 0 {
		return &dto.BatchProcessResponse{
			TotalVideos:  0,
			ErrorMessage: "未找到视频文件",
		}, nil
	}

	var processedVideos int
	var failedFiles []string

	// 处理每个视频文件
	for _, videoFile := range videoFiles {
		req := &dto.ProcessVideoRequest{InputPath: videoFile}
		resp, err := uc.ProcessSingleVideo(ctx, req)
		if err != nil {
			failedFiles = append(failedFiles, videoFile)
			log.Printf("处理视频失败: %s, 错误: %v", videoFile, err)
			continue
		}

		if resp.Success {
			processedVideos++
		} else {
			failedFiles = append(failedFiles, videoFile)
			log.Printf("处理视频失败: %s, 错误: %s", videoFile, resp.ErrorMessage)
		}
	}

	return &dto.BatchProcessResponse{
		TotalVideos:     totalVideos,
		ProcessedVideos: processedVideos,
		FailedVideos:    len(failedFiles),
		FailedFiles:     failedFiles,
	}, nil
}
