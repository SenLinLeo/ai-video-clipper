package main

import (
	"fmt"
	"log"
	"os"

	"ai-video-clipper/internal/interface/cli"
)

const (
	defaultConfigPath = "config.json"
)

func main() {
	// 解析命令行参数
	inputPath, configPath := parseArgs()

	// 创建CLI
	videoCLI, err := cli.NewVideoCLI(configPath)
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}

	// 根据参数决定处理方式
	if inputPath != "" {
		// 处理单个视频
		if err := videoCLI.ProcessSingleVideo(inputPath); err != nil {
			log.Fatalf("处理视频失败: %v", err)
		}
	} else {
		// 批量处理
		if err := videoCLI.ProcessBatch(); err != nil {
			log.Fatalf("批量处理失败: %v", err)
		}
	}

	log.Println("处理完成!")
}

// parseArgs 解析命令行参数
func parseArgs() (inputPath, configPath string) {
	configPath = defaultConfigPath

	if len(os.Args) > 1 {
		arg1 := os.Args[1]
		// 如果第一个参数是.json文件，当作配置文件处理
		if len(arg1) > 5 && arg1[len(arg1)-5:] == ".json" {
			configPath = arg1
			inputPath = "" // 批量处理模式
		} else {
			inputPath = arg1 // 单个视频文件
		}
	}

	if len(os.Args) > 2 {
		configPath = os.Args[2]
	}

	// 显示使用说明
	if len(os.Args) == 1 {
		fmt.Println("AI Video Clipper - DDD架构版本")
		fmt.Println("用法:")
		fmt.Println("  批量处理: ./video-clipper [config.json]")
		fmt.Println("  单个文件: ./video-clipper <input_video> [config.json]")
		fmt.Println()
		fmt.Printf("使用配置文件: %s\n", configPath)
		fmt.Println()
	}

	return inputPath, configPath
}
