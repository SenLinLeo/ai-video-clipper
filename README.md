# 🎬 AI视频剪辑器 (面向对象版)

高质量视频剪辑和转换工具，支持多种分辨率和速度处理，采用面向对象设计架构，支持高并发批量处理。

## ✨ 特性

### 🎯 核心功能
- **批量处理**: 自动处理输入目录中的所有视频文件
- **4版本输出**: 每个视频生成4个不同版本，全部输出10秒视频
- **智能剪辑**: 支持多种剪辑策略（开始截取、结尾截取、中间截取等）
- **高质量转换**: 使用FFmpeg进行专业级视频处理，模拟Adobe Media Encoder
- **高并发处理**: 支持最多100个视频和400个配置同时处理

### 🏗️ 技术架构
- **面向对象设计**: 使用Go接口实现模块化架构
- **接口分离**: ConfigProvider、PathGenerator、VideoConverter、ClipCalculator等独立接口
- **并发处理**: 基于errgroup的高性能并发管理
- **可扩展性**: 易于添加新的处理策略和配置选项
- **错误处理**: 完善的错误处理和日志记录

## 📋 视频输出规格

### 🎯 4个版本详细说明

每个视频自动生成**4个版本**，全部输出**10秒视频**：

| 版本 | 尺寸 | 剪辑策略 | 说明 |
|------|------|----------|------|
| 1 | **1008×1008** | `start_segments` | 从开始第5秒截取20秒 (5s-25s) → 2x速 → 10s |
| 2 | **1008×762** | `start_segments` | 从开始第5秒截取20秒 (5s-25s) → 2x速 → 10s |
| 3 | **1008×1008** | `end_segments` | 从倒数第25秒截取20秒 (倒数25s-倒数5s) → 2x速 → 10s |
| 4 | **1008×762** | `end_segments` | 从倒数第25秒截取20秒 (倒数25s-倒数5s) → 2x速 → 10s |

### 🔧 技术参数
- **视频比特率**: 4Mbps VBR (模拟Adobe Media Encoder)
- **音频比特率**: 112kbps AAC
- **播放速度**: 统一2.0倍速
- **输出时长**: 统一10秒
- **编码格式**: H.264 High Profile Level 4.1
- **像素格式**: YUV420P (兼容性最佳)
- **音频采样率**: 48kHz

## 🎬 视频压缩算法优化

### 📋 优化概述

将原本的简单FFmpeg编码参数升级为智能化、分层次的高效压缩系统，在保持画质的前提下优化文件大小和编码速度。

### ✨ 核心优化特性

#### 1. 🎯 三层质量预设系统
- **Fast** (`fast`): 快速编码，适合大批量处理
- **High** (`high`): 高质量编码，平衡质量与速度 ⭐ **默认**
- **Ultra** (`ultra`): 超高质量，追求极致画质

#### 2. 🔄 两阶段编码策略
- **第一阶段（剪辑）**: 使用快速预设，保证处理速度
- **第二阶段（最终输出）**: 使用指定质量预设，保证最终质量

#### 3. 🧠 智能编码参数
##### Fast预设优化
```
- Preset: faster
- CRF: 23 (平衡质量和文件大小)
- B帧: 3个
- 参考帧: 2个
- 运动估计: hex (快速)
- 子像素精度: 6级
```

##### High预设优化
```
- Preset: slow
- CRF: 20 (高质量)
- B帧: 5个
- 参考帧: 5个
- 运动估计: umh (高精度)
- 子像素精度: 8级
```

##### Ultra预设优化
```
- Preset: veryslow
- CRF: 18 (超高质量)
- B帧: 8个
- 参考帧: 8个
- 运动估计: tesa (最高精度)
- 子像素精度: 10级
```

#### 4. 🎨 心理视觉优化
- **启用心理视觉RD优化**: 提升主观画质
- **心理视觉Trellis**: 减少视觉瑕疵
- **自适应去块滤波**: 消除块效应

#### 5. ⚡ 高级编码技术
- **CABAC熵编码**: 提高压缩效率
- **混合8x8 DCT**: 提升细节表现
- **Trellis量化**: 优化量化过程
- **宏块树分析**: 智能比特率分配

### 🎛️ 压缩质量配置

#### 在config.json中设置质量预设
```json
{
  "inputDir": "/Volumes/Data/youtube-download",
  "outputDir": "/Volumes/Data/output",
  "qualityPreset": "high",  // fast, high, ultra
  "maxConcurrentVideos": 10,
  "maxConcurrentConfigs": 50,
  "batchSize": 20
}
```

#### 智能预设选择逻辑
如果未指定`qualityPreset`，系统会根据批处理大小自动选择：
- 批量 > 50个: 自动选择`fast`
- 批量 10-50个: 自动选择`high`
- 批量 < 10个: 自动选择`ultra`

### 📊 性能对比

| 质量预设 | 编码速度 | 文件大小 | 主观质量 | 适用场景 |
|---------|---------|---------|---------|---------|
| Fast    | ⚡⚡⚡⚡⚡ | 📦📦📦   | ⭐⭐⭐   | 大批量处理 |
| High    | ⚡⚡⚡     | 📦📦     | ⭐⭐⭐⭐  | 平衡质量速度 |
| Ultra   | ⚡       | 📦       | ⭐⭐⭐⭐⭐ | 高质量精品 |

### 🔬 技术细节

#### CRF vs 比特率模式
- **第一阶段**: 使用比特率模式，保证编码速度和一致性
- **第二阶段**: 使用CRF模式，保证最佳视觉质量

#### x264高级选项
```bash
# 示例：High预设的x264选项
-x264opts bframes=5:ref=5:me=umh:subme=8:trellis=2:keyint=50:min-keyint=25:scenecut=40:psy-rd=1.0:0.1:psy-trellis=0.1:deblock=-1:-1:cabac=1:8x8dct=1
```

#### 码率控制优化
- **最大码率**: 目标码率的1.5倍（CRF模式）或1.2倍（比特率模式）
- **缓冲区**: 目标码率的2倍，保证稳定编码
- **GOP结构**: 50帧关键帧间隔，25帧最小间隔

### 🎯 质量保证

#### 画质无损优化
1. **保持原始色彩空间**: yuv420p
2. **高精度运动估计**: umh/tesa算法
3. **心理视觉优化**: 提升主观感受
4. **智能场景检测**: 自动处理场景切换

#### 兼容性保证
- **H.264 High Profile Level 4.1**: 广泛设备支持
- **AAC音频编码**: 保持音质和兼容性
- **FastStart优化**: 支持网络流播放

### 📈 批量处理推荐
- **大量视频(>100个)**: 使用`fast`预设
- **中等批量(20-100个)**: 使用`high`预设
- **精品制作(<20个)**: 使用`ultra`预设

### 🔧 自定义调优

如需进一步自定义，可以修改`main.go`中的`VideoQualityParams`结构体参数：

```go
// 示例：自定义超高质量预设
case QualityUltra:
    return VideoQualityParams{
        Preset:      "veryslow",
        CRF:         16,  // 更高质量
        BFrames:     10,  // 更多B帧
        RefFrames:   10,  // 更多参考帧
        // ... 其他参数
    }
```

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
├── 1008x1008_start/           # 正方形，开始剪辑
│   ├── video1_square_start.mp4
│   ├── video2_square_start.mp4
│   └── subfolder/
│       └── video3_square_start.mp4
├── 1008x762_start/            # 矩形，开始剪辑
│   ├── video1_rect_start.mp4
│   └── ...
├── 1008x1008_end/             # 正方形，结尾剪辑
│   ├── video1_square_end.mp4
│   └── ...
└── 1008x762_end/              # 矩形，结尾剪辑
    ├── video1_rect_end.mp4
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
# 编译DDD版本（推荐）
cd cmd/clipper
go build -o ../../video-clipper .
cd ../..
```

**注意**: 原版本已归档到`legacy/`目录，推荐使用DDD架构版本。

#### 3. 准备视频文件
```bash
# 将视频文件放入 input/ 目录
mkdir -p input
cp /path/to/your/videos/* input/
```

#### 4. 执行批处理

```bash
# 批量处理
./video-clipper

# 处理单个文件
./video-clipper input.mp4

# 使用自定义配置
./video-clipper input.mp4 custom-config.json
```

### 使用脚本

#### 一键设置环境
```bash
./scripts/setup.sh
```

#### 快速演示
```bash
./scripts/demo.sh
```

## ⚡ 并发处理

程序支持高性能多级并发：
- **视频级并发**: 最多同时处理100个视频
- **配置级并发**: 每个视频的4个版本同时处理（最多100个并发）
- **极致性能**: 适合大规模批量处理
- **自动优化**: 根据CPU核心数调整并发参数

⚠️ **注意**: 高并发会消耗大量系统资源，建议：
- CPU: 16核心以上推荐
- 内存: 32GB以上推荐
- 存储: 高速SSD必需

### 📊 处理流程

```
📁 扫描视频文件
    ↓
🔄 并发处理每个视频 (最多100个同时)
    ↓
⚡ 每个视频并发生成4个版本 (最多100个配置并发)
    ↓
📈 显示处理统计
```

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
2. **ConfigurableConfigProvider**: 配置管理，支持JSON配置文件
3. **DefaultPathGenerator**: 路径生成，管理输出目录结构
4. **FFmpegVideoConverter**: FFmpeg接口，处理视频转换
5. **DefaultClipCalculator**: 剪辑计算，实现不同剪辑策略
6. **errgroup**: 并发管理，支持高性能多协程处理

## ⚙️ 配置说明

### 自定义配置

编辑 `config.json` 文件自定义参数：

```json
{
  "inputDir": "input",
  "outputDir": "output",
  "audioBitrate": "112k",
  "qualityPreset": "high",
  "maxConcurrentVideos": 10,
  "maxConcurrentConfigs": 50,
  "batchSize": 20,
  "videoConfigs": [
    {
      "Width": 1008,
      "Height": 1008,
      "ClipDuration": 20,
      "Speed": 2.0,
      "VideoBitrate": 4000,
      "ClipStrategy": "start_segments",
      "OutputSuffix": "_square_start",
      "OutputFolder": "1008x1008_start"
    }
    // ... 更多配置
  ]
}
```

### Config 完整结构
```go
type Config struct {
    InputDir             string        // 输入目录
    OutputDir            string        // 输出目录
    AudioBitrate         string        // 音频比特率
    QualityPreset        string        // 质量预设: fast, high, ultra
    MaxConcurrentVideos  int           // 最大并发视频数
    MaxConcurrentConfigs int           // 最大并发配置数
    BatchSize            int           // 批处理大小
    VideoConfigs         []VideoConfig // 视频配置列表
}

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

## 🎯 剪辑策略详解

### Start Segments (开始剪辑)
- **用途**: 版本1和2
- **逻辑**: 从开始第5秒开始截取20秒内容 (5s-25s)
- **处理**: 20秒原片段 → 2.0倍速 → **10秒输出**

### End Segments (结尾剪辑)
- **用途**: 版本3和4
- **逻辑**: 从倒数第25秒开始截取20秒内容 (倒数25s-倒数5s)
- **处理**: 20秒原片段 → 2.0倍速 → **10秒输出**

### Middle Segments (中间剪辑)
- **逻辑**: 从视频前5秒后到后5秒前的中间部分截取指定时长
- **处理**: N秒原片段 → 速度调整 → 10秒输出
- **公式**: 可用时间 = (视频长度 - 10秒)，从中间截取

### Last Segments (尾部剪辑)
- **逻辑**: 从视频结束前5秒往前截取指定时长内容
- **处理**: N秒原片段 → 速度调整 → 10秒输出

## 📈 性能监控

### 运行输出示例
```
📄 使用默认配置
🎬 视频批处理器启动 (面向对象版)
==============================
输入目录: input
输出目录: output
音频比特率: 112k
支持格式: [.mp4 .avi .mov .mkv .flv .wmv .m4v .3gp .webm]

📋 处理配置:
  1. 1008x1008_start: 1008x1008, 2.0x速, start_segments策略, 4000k比特率
  2. 1008x762_start: 1008x762, 2.0x速, start_segments策略, 4000k比特率
  3. 1008x1008_end: 1008x1008, 2.0x速, end_segments策略, 4000k比特率
  4. 1008x762_end: 1008x762, 2.0x速, end_segments策略, 4000k比特率
==============================

📁 找到 3 个视频文件，开始并发处理...

🎬 开始处理视频: input/video.mp4
  [1/4] 生成 1008x1008_start 版本...
  [2/4] 生成 1008x762_start 版本...
  [3/4] 生成 1008x1008_end 版本...
  [4/4] 生成 1008x762_end 版本...
✅ 视频 input/video.mp4 处理完成

📊 处理统计: 成功 3 个，失败 0 个，总计 3 个视频

🎉 所有视频处理完成!
```

## 📊 支持格式

### 输入格式
`.mp4`, `.avi`, `.mov`, `.mkv`, `.flv`, `.wmv`, `.m4v`, `.3gp`, `.webm`

### 输出格式
统一输出为 `.mp4` 格式，兼容性最佳

## 🎯 最佳实践

1. **视频要求**: 建议输入视频长度 > 30秒
2. **硬件配置**:
   - CPU: 16核心以上推荐（高并发）
   - 内存: 32GB以上推荐（高并发）
   - 存储: 高速SSD必需
3. **批量处理**: 支持子目录，自动保持目录结构
4. **并发控制**: 可根据硬件性能调整并发数

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

#### 3. 高并发相关问题
- **内存不足**: 减少并发数或增加系统内存
- **CPU过载**: 调整并发参数或使用更快的预设
- **磁盘I/O瓶颈**: 使用高速SSD存储

#### 4. 视频处理问题
- **视频太短**: 确保输入视频 > 30秒
- **输出文件异常小**: 检查剪辑时间计算和FFmpeg参数
- **质量问题**: 调整 `VideoBitrate` 参数

### 调试模式
```bash
# 查看详细 FFmpeg 输出
FFMPEG_DEBUG=1 ./video-clipper
```

## 📈 性能优化

### 处理速度优化
- 使用高速SSD存储
- 增加系统内存
- 使用多核CPU
- 调整并发参数

### 质量优化
- 调整视频比特率
- 使用不同的预设参数
- 优化编码设置

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

### 调整并发参数
```go
// 在 main.go 中修改并发常量
const (
    maxConcurrentVideos  = 50  // 调整视频并发数
    maxConcurrentConfigs = 50  // 调整配置并发数
)
```

## 🏗️ DDD架构重构

### 重构概述

本项目已使用领域驱动设计(DDD)思想进行重构，保持最小改动的同时提升了代码的可维护性和可扩展性。

### 架构变化

#### 原架构（已归档）
```
legacy/
├── main_legacy.go (1070行)   # 原版本主程序（已归档）
├── errgroup_legacy.go        # 并发控制（已归档）
└── config.go                 # 配置管理
```

#### 新DDD架构
```
cmd/clipper/main.go                          # 应用入口
internal/
├── domain/                                  # 领域层
│   ├── entities/                           # 实体
│   │   ├── video.go                        # 视频实体
│   │   └── processing_config.go            # 处理配置实体
│   ├── repositories/                       # 仓储接口
│   │   ├── video_repository.go             # 视频仓储接口
│   │   └── config_repository.go            # 配置仓储接口
│   └── services/                           # 领域服务
│       └── path_service.go                 # 路径服务
├── application/                            # 应用层
│   ├── dto/                                # 数据传输对象
│   │   └── video_processing_dto.go         # 视频处理DTO
│   └── usecases/                           # 用例
│       └── video_processing_usecase.go     # 视频处理用例
├── infrastructure/                         # 基础设施层
│   ├── config/                             # 配置实现
│   │   └── config_repository_impl.go       # 配置仓储实现
│   └── ffmpeg/                             # FFmpeg实现
│       └── video_repository_impl.go        # 视频仓储实现
└── interface/                              # 接口层
    └── cli/                                # CLI接口
        └── video_cli.go                    # CLI实现
```

### DDD核心改进

#### 1. 关注点分离
- **领域层**: 纯业务逻辑，不依赖任何技术实现
- **应用层**: 编排业务流程，定义用例
- **基础设施层**: 技术实现细节（FFmpeg、文件I/O）
- **接口层**: 外部交互适配器

#### 2. 依赖倒置
- 领域层定义接口，基础设施层实现接口
- 高层模块不依赖低层模块，都依赖抽象

#### 3. 单一职责
- 每个类/函数都有明确的单一职责
- 视频实体只关心视频属性和验证
- 配置实体只关心配置验证和计算
- 仓储只关心数据访问

#### 4. 可测试性
- 通过接口注入依赖，便于单元测试
- 业务逻辑与技术实现分离

### DDD版本使用方式

#### 编译和运行
```bash
# 编译
cd cmd/clipper
go build -o ../../video-clipper .
cd ../..

# 批量处理
./video-clipper

# 处理单个文件
./video-clipper input.mp4

# 使用自定义配置
./video-clipper input.mp4 custom-config.json
```

### 兼容性保证

- 完全兼容原有的配置文件格式
- 保持相同的命令行接口
- 输出结果与原版本一致

### 扩展性提升

新DDD架构便于扩展：

1. **新的视频格式支持**: 在`Video`实体中添加格式验证
2. **新的处理策略**: 在`ProcessingConfig`实体中添加策略
3. **新的输出格式**: 实现新的`VideoRepository`
4. **新的配置源**: 实现新的`ConfigRepository`
5. **新的接口**: 添加HTTP API、gRPC等接口层

### DDD性能特性

- 保持原有的并发处理能力
- 通过接口抽象，便于性能优化
- 可独立优化各层实现

### DDD代码质量

- 函数长度控制在50行以内
- 遵循SOLID原则
- 清晰的错误处理
- 完整的类型安全

### 版本说明

项目已完全迁移到DDD架构：
- **当前版本**: `video-clipper` - DDD重构版本（领域驱动设计）
- **归档版本**: `legacy/` - 原版本（面向对象设计，已归档）

推荐使用DDD版本，具有更好的可维护性和可扩展性。

## 📝 变更日志

### v4.1.0 (架构简化版) - 2024-01-06
- 🎯 简化为单一DDD版本，原版本归档到`legacy/`
- 🏗️ 统一使用`video-clipper`作为可执行文件名
- 📁 重新组织目录结构：`scripts/`存放脚本，`legacy/`存放归档文件
- 🔧 更新所有脚本和文档，指向DDD版本

### v4.0.0 (DDD重构版) - 2024-01-06
- 🏗️ 使用DDD思想重构整体架构
- 🎯 四层架构：领域层、应用层、基础设施层、接口层
- 🔧 依赖倒置和关注点分离
- ⚡ 保持原有高并发处理能力
- 📊 提升代码可维护性和可扩展性

### v3.0.0 (高并发版) - 2024-01-06
- 🚀 增加高并发处理支持（最多100个视频同时处理）
- 🎯 重新设计4个版本配置（开始剪辑 + 结尾剪辑）
- 🔧 Adobe Media Encoder风格的VBR编码
- ⚡ 基于errgroup的并发管理
- 📊 增强的处理统计和监控

### v2.0.0 (面向对象版) - 2024-01-06
- ✨ 完全重构为面向对象架构
- 🏗️ 引入接口分离原则
- 📂 支持JSON配置文件
- 🔧 分步处理机制提升稳定性
- 🎯 改进的剪辑策略算法
- 📊 增强的错误处理和日志

## 📄 许可证

MIT License - 详见 LICENSE 文件

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

**项目地址**: `ai-video-clipper/`
**技术栈**: Go + FFmpeg + errgroup并发
**架构模式**: 面向对象 + 接口设计 + 高并发处理 + DDD领域驱动设计