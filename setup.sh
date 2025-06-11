#!/bin/bash

echo "🎬 视频剪辑器安装脚本"
echo "====================="

# 检查FFmpeg是否已安装
if command -v ffmpeg &> /dev/null; then
    echo "✅ FFmpeg已安装"
    ffmpeg -version | head -1
else
    echo "📦 正在安装FFmpeg..."
    if command -v brew &> /dev/null; then
        brew install ffmpeg
        echo "✅ FFmpeg安装完成"
    else
        echo "❌ 错误: 未找到Homebrew，请手动安装FFmpeg"
        echo "   访问: https://ffmpeg.org/download.html"
        exit 1
    fi
fi

# 编译程序
echo "🔨 正在编译视频剪辑器..."
go build -o video-clipper .

if [ $? -eq 0 ]; then
    echo "✅ 编译成功"
    echo ""
    echo "🚀 使用方法:"
    echo "1. 将视频文件放入 input/ 文件夹"
    echo "2. 运行: ./video-clipper"
    echo "3. 剪辑后的视频将保存在 output/ 文件夹"
    echo ""
    echo "📚 更多详情请查看 README.md"
else
    echo "❌ 编译失败"
    exit 1
fi 