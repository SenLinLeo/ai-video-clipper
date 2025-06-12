package cli

import (
	"context"
	"fmt"
	"log"
	"os/exec"

	"ai-video-clipper/internal/application/dto"
	"ai-video-clipper/internal/application/usecases"
	"ai-video-clipper/internal/domain/services"
	"ai-video-clipper/internal/infrastructure/config"
	"ai-video-clipper/internal/infrastructure/ffmpeg"
)

// VideoCLI 视频处理CLI
type VideoCLI struct {
	useCase *usecases.VideoProcessingUseCase
}

// NewVideoCLI 创建视频处理CLI
func NewVideoCLI(configPath string) (*VideoCLI, error) {
	// 检查FFmpeg
	if err := checkFFmpeg(); err != nil {
		return nil, err
	}

	// 创建配置仓储
	configRepo := config.NewConfigRepositoryImpl(configPath)

	// 获取全局配置
	ctx := context.Background()
	globalConfig, err := configRepo.GetGlobalConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取全局配置失败: %v", err)
	}

	// 创建视频仓储
	videoRepo := ffmpeg.NewVideoRepositoryImpl(globalConfig)

	// 创建路径服务
	pathService := services.NewPathService(globalConfig.InputDir, globalConfig.OutputDir)

	// 创建用例
	useCase := usecases.NewVideoProcessingUseCase(videoRepo, configRepo, pathService)

	return &VideoCLI{
		useCase: useCase,
	}, nil
}

// ProcessSingleVideo 处理单个视频
func (cli *VideoCLI) ProcessSingleVideo(inputPath string) error {
	ctx := context.Background()

	req := &dto.ProcessVideoRequest{
		InputPath: inputPath,
	}

	resp, err := cli.useCase.ProcessSingleVideo(ctx, req)
	if err != nil {
		return fmt.Errorf("处理视频失败: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("处理视频失败: %s", resp.ErrorMessage)
	}

	log.Printf("成功处理视频，生成了 %d 个文件", len(resp.ProcessedFiles))
	for _, file := range resp.ProcessedFiles {
		log.Printf("  - %s", file)
	}

	return nil
}

// ProcessBatch 批量处理视频
func (cli *VideoCLI) ProcessBatch() error {
	ctx := context.Background()

	req := &dto.BatchProcessRequest{}

	resp, err := cli.useCase.ProcessBatch(ctx, req)
	if err != nil {
		return fmt.Errorf("批量处理失败: %v", err)
	}

	if resp.ErrorMessage != "" {
		log.Printf("处理过程中出现错误: %s", resp.ErrorMessage)
	}

	log.Printf("批量处理完成:")
	log.Printf("  总视频数: %d", resp.TotalVideos)
	log.Printf("  成功处理: %d", resp.ProcessedVideos)
	log.Printf("  处理失败: %d", resp.FailedVideos)

	if len(resp.FailedFiles) > 0 {
		log.Printf("失败的文件:")
		for _, file := range resp.FailedFiles {
			log.Printf("  - %s", file)
		}
	}

	return nil
}

// DisplayConfig 显示配置信息
func (cli *VideoCLI) DisplayConfig() error {
	// 这里可以添加配置显示逻辑
	log.Printf("配置信息显示功能待实现")

	return nil
}

// checkFFmpeg 检查FFmpeg是否可用
func checkFFmpeg() error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("未找到 ffmpeg，请确保已安装 FFmpeg 并添加到 PATH")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return fmt.Errorf("未找到 ffprobe，请确保已安装 FFmpeg 并添加到 PATH")
	}
	return nil
}
