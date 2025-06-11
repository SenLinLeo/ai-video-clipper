#!/bin/bash

echo "🎬 视频剪辑器演示"
echo "================="

# 检查是否已编译
if [ ! -f "video-clipper" ]; then
    echo "❌ 程序未编译，请先运行: ./setup.sh"
    exit 1
fi

# 检查配置文件是否存在
if [ ! -f "config.json" ]; then
    echo "❌ 配置文件 config.json 不存在"
    exit 1
fi

# 从配置文件读取输入和输出目录
input_dir=$(grep '"inputDir"' config.json | sed 's/.*"inputDir": *"\([^"]*\)".*/\1/')
output_dir=$(grep '"outputDir"' config.json | sed 's/.*"outputDir": *"\([^"]*\)".*/\1/')

echo "📁 配置文件设置:"
echo "  输入目录: $input_dir"
echo "  输出目录: $output_dir"
echo ""

# 创建目录结构
echo "📁 确保目录存在..."
mkdir -p "$input_dir"
mkdir -p "$output_dir"

# 检查输入目录是否有视频文件
video_count=$(find "$input_dir" -type f \( -iname "*.mp4" -o -iname "*.avi" -o -iname "*.mov" -o -iname "*.mkv" -o -iname "*.flv" -o -iname "*.wmv" -o -iname "*.m4v" -o -iname "*.3gp" -o -iname "*.webm" \) | wc -l)

if [ $video_count -eq 0 ]; then
    echo "📝 输入目录中没有发现视频文件"
    echo ""
    echo "请将视频文件放入: $input_dir"
    echo ""
    echo "支持的格式: mp4, avi, mov, mkv, flv, wmv, m4v, 3gp, webm"
    exit 1
else
    echo "✅ 发现 $video_count 个视频文件"
fi

# 运行视频剪辑器
echo ""
echo "🚀 开始处理视频..."
echo "==================="
./video-clipper config.json

echo ""
echo "🎉 处理完成！"
echo ""

# 显示结果
if [ -d "$output_dir" ] && [ "$(ls -A "$output_dir")" ]; then
    echo "📁 输出文件:"
    find "$output_dir" -type f \( -name "*_square_start.*" -o -name "*_rect_start.*" -o -name "*_square_end.*" -o -name "*_rect_end.*" \) | sort
else
    echo "⚠️ 输出目录为空，可能处理过程中出现了错误"
fi 