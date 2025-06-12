#!/bin/bash

# 文件拆分脚本
# 用法: ./split_files.sh <源目录> [每批文件数量]

set -e  # 遇到错误立即退出

# 默认参数
DEFAULT_BATCH_SIZE=100

# 检查参数
if [ $# -lt 1 ]; then
    echo "用法: $0 <源目录> [每批文件数量]"
    echo "示例: $0 /Volumes/Data/youtube-download 100"
    exit 1
fi

SOURCE_DIR="$1"
BATCH_SIZE="${2:-$DEFAULT_BATCH_SIZE}"

# 检查源目录是否存在
if [ ! -d "$SOURCE_DIR" ]; then
    echo "❌ 错误: 源目录不存在: $SOURCE_DIR"
    exit 1
fi

# 获取源目录的基本名称（去掉路径）
BASE_NAME=$(basename "$SOURCE_DIR")
PARENT_DIR=$(dirname "$SOURCE_DIR")

echo "📁 文件拆分器"
echo "=============="
echo "源目录: $SOURCE_DIR"
echo "目标基本名: $BASE_NAME"
echo "每批文件数: $BATCH_SIZE"
echo ""

# 计算文件总数（不包括子目录）
echo "📊 统计文件数量..."
TOTAL_FILES=$(find "$SOURCE_DIR" -maxdepth 1 -type f | wc -l | tr -d ' ')

if [ "$TOTAL_FILES" -eq 0 ]; then
    echo "⚠️ 源目录中没有找到文件"
    exit 0
fi

echo "✅ 找到 $TOTAL_FILES 个文件"

# 计算需要的批次数
TOTAL_BATCHES=$(( (TOTAL_FILES + BATCH_SIZE - 1) / BATCH_SIZE ))
echo "📦 将创建 $TOTAL_BATCHES 个批次目录"

# 确认操作
echo ""
read -p "确认继续操作? (y/N): " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ 操作已取消"
    exit 0
fi

echo ""
echo "🚀 开始拆分文件..."

# 初始化计数器
current_batch=1
files_in_current_batch=0
current_target_dir=""

# 创建目标目录的函数
create_target_dir() {
    local batch_num=$1
    local target_dir="$PARENT_DIR/${BASE_NAME}-${batch_num}"
    
    if [ -d "$target_dir" ]; then
        echo "⚠️ 目录已存在，清空并继续: $target_dir"
        rm -rf "$target_dir"
    fi
    
    mkdir -p "$target_dir"
    echo "📁 创建目录: $target_dir"
    echo "$target_dir"
}

# 处理文件
file_count=0
find "$SOURCE_DIR" -maxdepth 1 -type f -print0 | while IFS= read -r -d '' file; do
    # 如果当前批次已满或者是第一个文件，创建新的目标目录
    if [ $files_in_current_batch -eq 0 ]; then
        current_target_dir=$(create_target_dir $current_batch)
    fi
    
    # 移动文件
    filename=$(basename "$file")
    mv "$file" "$current_target_dir/"
    
    file_count=$((file_count + 1))
    files_in_current_batch=$((files_in_current_batch + 1))
    
    # 显示进度
    if [ $((file_count % 10)) -eq 0 ]; then
        echo "📄 已处理: $file_count/$TOTAL_FILES 文件 (当前批次: $current_batch, 批次内: $files_in_current_batch/$BATCH_SIZE)"
    fi
    
    # 如果当前批次已满，准备下一批次
    if [ $files_in_current_batch -eq $BATCH_SIZE ]; then
        echo "✅ 批次 $current_batch 完成: $files_in_current_batch 个文件"
        current_batch=$((current_batch + 1))
        files_in_current_batch=0
    fi
done

echo ""
echo "🎉 文件拆分完成！"
echo ""

# 显示结果统计
echo "📊 拆分结果："
for i in $(seq 1 $TOTAL_BATCHES); do
    target_dir="$PARENT_DIR/${BASE_NAME}-${i}"
    if [ -d "$target_dir" ]; then
        file_count_in_dir=$(find "$target_dir" -maxdepth 1 -type f | wc -l | tr -d ' ')
        echo "  $target_dir: $file_count_in_dir 个文件"
    fi
done

# 检查原目录是否为空
remaining_files=$(find "$SOURCE_DIR" -maxdepth 1 -type f | wc -l | tr -d ' ')
if [ "$remaining_files" -eq 0 ]; then
    echo ""
    echo "✅ 原目录已清空，可以安全删除"
    read -p "是否删除空的原目录? (y/N): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rmdir "$SOURCE_DIR"
        echo "🗑️ 已删除原目录: $SOURCE_DIR"
    fi
else
    echo "⚠️ 原目录中还有 $remaining_files 个文件未处理"
fi

echo ""
echo "✅ 所有操作完成！" 