# 🎬 视频批处理器 (面向对象版)

高质量视频剪辑和转换工具，支持多种分辨率和速度处理，采用面向对象设计架构。

## ✨ 特性

### 🎯 核心功能
- **批量处理**: 自动处理输入目录中的所有视频文件
- **多版本输出**: 每个视频生成4个不同版本
- **智能剪辑**: 支持两种剪辑策略（尾部截取、中间截取）
- **高质量转换**: 使用FFmpeg进行专业级视频处理

### 🏗️ 技术架构
- **面向对象设计**: 使用Go接口实现模块化架构
- **接口分离**: ConfigProvider、PathGenerator、VideoConverter、ClipCalculator等独立接口
- **可扩展性**: 易于添加新的处理策略和配置选项
- **错误处理**: 完善的错误处理和日志记录

### 📐 输出规格

#### 四种版本配置：
1. **1008x1008_2.5x**: 正方形，2.5倍速，尾部剪辑策略
2. **1008x762_2.5x**: 矩形，2.5倍速，尾部剪辑策略
3. **1008x1008_2.0x**: 正方形，2.0倍速，中间剪辑策略
4. **1008x762_2.0x**: 矩形，2.0倍速，中间剪辑策略

#### 技术参数：
- **音频比特率**: 112kbps AAC
- **视频编码**: H.264 High Profile Level 4.1
- **像素格式**: YUV420P (兼容性最佳)
- **压缩质量**: CRF 20 (高质量)

## 🔧 架构说明

### 接口设计

```go
// 主要接口
type VideoProcessor interface {
    ProcessAllVideos() error
    ProcessVideo(inputPath string) error
}

type ConfigProvider interface {
    GetConfigs() []VideoConfig
    ValidateConfig(config VideoConfig) error
}

type PathGenerator interface {
    GenerateOutputPath(inputPath string, config VideoConfig) (string, error)
    EnsureDir(dir string) error
}

type VideoConverter interface {
    ConvertVideo(inputPath, outputPath string, config VideoConfig) error
    GetVideoDuration(videoPath string) (float64, error)
}

type ClipCalculator interface {
    CalculateClipTimes(totalDuration float64, config VideoConfig) (float64, float64, error)
}
```

### 核心组件

1. **DefaultVideoProcessor**: 主处理器，协调所有组件
2. **DefaultConfigProvider**: 配置管理，提供四种预设配置
3. **DefaultPathGenerator**: 路径生成，管理输出目录结构
4. **FFmpegVideoConverter**: FFmpeg接口，处理视频转换
5. **DefaultClipCalculator**: 剪辑计算，实现不同剪辑策略

## 📂 目录结构

### 输入结构
```
input/
├── video1.mp4
├── video2.avi
└── subfolder/
    └── video3.mov
```

### 输出结构
```
output/
├── 1008x1008_2.5x/          # 正方形 2.5倍速
│   ├── video1_square_last.mp4
│   ├── video2_square_last.mp4
│   └── subfolder/
│       └── video3_square_last.mp4
├── 1008x762_2.5x/           # 矩形 2.5倍速
│   ├── video1_rect_last.mp4
│   └── ...
├── 1008x1008_2.0x/          # 正方形 2.0倍速
│   ├── video1_square_middle.mp4
│   └── ...
└── 1008x762_2.0x/           # 矩形 2.0倍速
    ├── video1_rect_middle.mp4
    └── ...
```

## 🚀 安装使用

### 前提条件
- Go 1.19+ 
- FFmpeg (包含ffprobe)

### 快速开始

#### 1. 安装依赖 (macOS)
```bash
brew install ffmpeg
```

#### 2. 编译程序
```bash
go build -o video-clipper .
```

#### 3. 准备视频文件
```bash
# 将视频文件放入 input/ 目录
mkdir -p input
cp /path/to/your/videos/* input/
```

#### 4. 执行批处理
```bash
# 使用默认目录 (input -> output)
./video-clipper

# 自定义目录
./video-clipper /path/to/input /path/to/output
```

### 使用脚本

#### 一键设置环境
```bash
./setup.sh
```

#### 快速演示
```bash
./demo.sh
```

## 🎯 剪辑策略详解

### Last Segments (尾部剪辑)
- **用途**: 2.5x 速度版本
- **逻辑**: 从视频结束前5秒往前截取25秒内容
- **公式**: 结束时间 = 视频长度 - 5秒，开始时间 = 结束时间 - 25秒
- **输出**: 经过2.5倍速播放后得到10秒视频

### Middle Segments (中间剪辑)  
- **用途**: 2.0x 速度版本
- **逻辑**: 从视频前5秒后到后5秒前的中间部分截取20秒
- **公式**: 可用时间 = (视频长度 - 10秒)，从中间截取20秒
- **输出**: 经过2.0倍速播放后得到10秒视频

## 🔧 处理流程

### 分步处理机制
1. **第一步**: 剪辑和缩放
   - 时间截取 (`-ss` 和 `-t` 参数)
   - 分辨率转换 (缩放和裁剪)
   - 基础编码 (中等质量)

2. **第二步**: 速度调整和优化
   - 视频调速 (`setpts` 滤镜)
   - 音频调速 (`atempo` 滤镜)
   - 高质量编码 (CRF 20)

### 音频处理特别说明
- 2.0x速度: 直接使用 `atempo=2.0`
- 2.5x速度: 分级处理 `atempo=2.0,atempo=1.25` (避免FFmpeg限制)

## 📊 支持格式

### 输入格式
`.mp4`, `.avi`, `.mov`, `.mkv`, `.flv`, `.wmv`, `.m4v`, `.3gp`, `.webm`

### 输出格式
统一输出为 `.mp4` 格式，兼容性最佳

## ⚙️ 配置说明

### VideoConfig 结构
```go
type VideoConfig struct {
    Width          int     // 视频宽度
    Height         int     // 视频高度  
    ClipDuration   int     // 剪辑时长（秒）
    Speed          float64 // 播放速度倍数
    VideoBitrate   int     // 视频比特率(kbps)
    ClipStrategy   string  // 剪辑策略
    OutputSuffix   string  // 输出文件后缀
    OutputFolder   string  // 输出文件夹
}
```

### 自定义配置
要添加新的处理配置，修改 `DefaultConfigProvider.GetConfigs()` 方法：

```go
func (dcp *DefaultConfigProvider) GetConfigs() []VideoConfig {
    return []VideoConfig{
        // 现有配置...
        {
            Width:        720,
            Height:       720,
            ClipDuration: 15,
            Speed:        1.5,
            VideoBitrate: 800,
            ClipStrategy: "last_segments",
            OutputSuffix: "_custom",
            OutputFolder: "720x720_1.5x",
        },
        // 更多配置...
    }
}
```

## 🚦 状态指示

### 运行输出示例
```
🎬 视频批处理器启动 (面向对象版)
==============================
输入目录: input
输出目录: output
音频比特率: 112k
支持格式: [.mp4 .avi .mov .mkv .flv .wmv .m4v .3gp .webm]

📋 处理配置:
  1. 1008x1008_2.5x: 1008x1008, 2.5x速, last_segments策略
  2. 1008x762_2.5x: 1008x762, 2.5x速, last_segments策略
  3. 1008x1008_2.0x: 1008x1008, 2.0x速, middle_segments策略
  4. 1008x762_2.0x: 1008x762, 2.0x速, middle_segments策略
==============================

🎬 开始处理视频: input/example.mp4
  [1/4] 生成 1008x1008_2.5x 版本...
  ✅ 成功生成 [1008x1008_2.5x]: output/1008x1008_2.5x/example_square_last.mp4
  [2/4] 生成 1008x762_2.5x 版本...
  ✅ 成功生成 [1008x762_2.5x]: output/1008x762_2.5x/example_rect_last.mp4
  [3/4] 生成 1008x1008_2.0x 版本...
  ✅ 成功生成 [1008x1008_2.0x]: output/1008x1008_2.0x/example_square_middle.mp4
  [4/4] 生成 1008x762_2.0x 版本...
  ✅ 成功生成 [1008x762_2.0x]: output/1008x762_2.0x/example_rect_middle.mp4
✅ 视频 input/example.mp4 处理完成

🎉 所有视频处理完成!
```

## 🐛 故障排除

### 常见问题

#### 1. FFmpeg 未找到
```bash
# macOS
brew install ffmpeg

# Ubuntu/Debian  
sudo apt update && sudo apt install ffmpeg

# CentOS/RHEL
sudo yum install ffmpeg
```

#### 2. 权限问题
```bash
chmod +x video-clipper
chmod +x *.sh
```

#### 3. 内存不足
- 减少并发处理
- 使用 `-preset fast` 而非 `slow`
- 降低输出质量设置

#### 4. 输出文件异常小 (334B)
- 检查输入视频时长是否足够
- 验证剪辑时间计算是否正确
- 查看 FFmpeg 错误日志

### 调试模式
```bash
# 查看详细 FFmpeg 输出
FFMPEG_DEBUG=1 ./video-clipper
```

## 📈 性能优化

### 处理速度优化
- 使用 SSD 存储
- 增加系统内存
- 使用 `-preset faster` (质量会稍微降低)

### 质量优化
- 使用 `-preset slow` (处理时间更长)
- 调整 CRF 值 (更低 = 更高质量)
- 增加视频比特率

## 🔮 扩展开发

### 添加新的剪辑策略
```go
// 在 ClipCalculator 中添加新策略
case "custom_strategy":
    // 自定义剪辑逻辑
    startTime = customCalculation(totalDuration, config)
    endTime = startTime + float64(config.ClipDuration)
```

### 添加新的输出格式
```go
// 在 VideoConverter 中添加编码选项
args = append(args, "-c:v", "libx265") // HEVC编码
```

### 添加进度回调
```go
type ProgressCallback func(current, total int, status string)

// 在 VideoProcessor 接口中添加
ProcessAllVideosWithProgress(callback ProgressCallback) error
```

## 📝 变更日志

### v2.0.0 (面向对象版) - 2024-01-06
- ✨ 完全重构为面向对象架构
- 🏗️ 引入接口分离原则
- 📂 简化输出目录命名 (分辨率_倍速)
- 🔧 分步处理机制提升稳定性
- 🎯 改进的剪辑策略算法
- 📊 增强的错误处理和日志

### v1.x.x (过往版本)
- 详见 CHANGELOG.md

## 📄 许可证

MIT License - 详见 LICENSE 文件

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

**项目地址**: `gpt-4o-image/video-clipper/`
**技术栈**: Go + FFmpeg
**架构模式**: 面向对象 + 接口设计 