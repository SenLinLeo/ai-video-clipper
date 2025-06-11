#!/bin/bash

echo "🎬 视频剪辑器演示"
echo "================="

# 检查是否已编译
if [ ! -f "video-clipper" ]; then
    echo "❌ 程序未编译，请先运行: ./setup.sh"
    exit 1
fi

# 创建示例目录结构
echo "📁 创建示例目录结构..."
mkdir -p input/subfolder
mkdir -p output

# 检查input目录是否有视频文件
video_count=$(find input -type f \( -iname "*.mp4" -o -iname "*.avi" -o -iname "*.mov" -o -iname "*.mkv" \) | wc -l)

if [ $video_count -eq 0 ]; then
    echo "📝 input目录中没有发现视频文件"
    echo ""
    echo "请将视频文件放入以下位置："
    echo "  input/video1.mp4"
    echo "  input/video2.avi"
    echo "  input/subfolder/video3.mov"
    echo ""
    echo "然后重新运行此脚本"
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
./video-clipper

echo ""
echo "🎉 处理完成！"
echo ""

# 显示结果
if [ -d "output" ] && [ "$(ls -A output)" ]; then
    echo "📁 输出文件:"
    find output -type f \( -name "*_last10s.mp4" -o -name "*_last10s.avi" -o -name "*_last10s.mov" -o -name "*_last10s.mkv" -o -name "*_last10s.flv" -o -name "*_last10s.wmv" -o -name "*_last10s.m4v" -o -name "*_last10s.3gp" -o -name "*_last10s.webm" \) | sort
else
    echo "⚠️ output目录为空，可能处理过程中出现了错误"
fi 