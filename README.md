# ffmcli

A high-performance, hardware-accelerated video transcoding tool built with Go. Optimized for both NVIDIA GPUs (NVENC) and Apple Silicon (VideoToolbox). Inspired by Shutter Encoder but designed as a streamlined command-line tool.

## 🚀 Features

- **Multi-Platform Hardware Acceleration**: 
  - NVIDIA GPUs: Utilizes NVENC for H.264, H.265, and AV1 encoding
  - Apple Silicon: Utilizes VideoToolbox for H.264 and H.265 encoding, plus optimized AV1 software encoding
- **Preset-Based System**: 7 optimized encoding presets for common scenarios
- **Parallel Processing**: Process multiple files simultaneously with configurable workers
- **Recursive Directory Processing**: Scan and process entire directory trees
- **Multiple Output Formats**: Support for various codecs and resolutions
- **Cross-Platform**: Runs on Linux, Windows, and macOS (with platform-specific optimizations)

## 📋 Prerequisites

### System Requirements

#### For NVIDIA GPU Systems:
- NVIDIA GPU with NVENC support (GTX 10 series or newer recommended)
- NVIDIA drivers (latest recommended)
- FFmpeg with NVIDIA support
- CUDA toolkit (recommended for optimal performance)

#### For Apple Silicon Macs:
- Apple Silicon Mac (M1, M2, M3, or newer)
- macOS 11.0 or later
- FFmpeg with VideoToolbox support (`brew install ffmpeg`)
- For best AV1 support: FFmpeg with SVT-AV1 encoder

#### For Other Systems:
- FFmpeg with appropriate codec support
- Software encoding fallback available

### Supported Input Formats
MP4, MKV, AVI, MOV, WMV, FLV, WebM, M4V, 3GP, TS, MTS, M2TS

## 🛠️ Installation

### Option 1: Download Pre-built Binary
Download the latest release from the [Releases](https://github.com/hemp0r/ffmcli/releases) page.

### Option 2: Build from Source
```bash
git clone https://github.com/hemp0r/ffmcli.git
cd ffmcli/go
make build
```

### FFmpeg Installation

#### macOS (Apple Silicon)
```bash
# Install FFmpeg with VideoToolbox and SVT-AV1 support
brew install ffmpeg

# For advanced AV1 encoding, install with SVT-AV1
brew install ffmpeg --with-svt-av1  # if available, or:
brew install svt-av1
```

#### Ubuntu/Debian (NVIDIA)
```bash
# Add NVIDIA codec repositories
sudo apt update
sudo apt install ffmpeg nvidia-driver nvidia-cuda-toolkit

# For latest FFmpeg with full NVENC support
sudo add-apt-repository ppa:graphics-drivers/ppa
sudo apt update && sudo apt install ffmpeg
```

#### Windows (NVIDIA)
1. Install [NVIDIA drivers](https://www.nvidia.com/Download/index.aspx)
2. Download [FFmpeg](https://ffmpeg.org/download.html) with NVENC support
3. Add FFmpeg to your PATH

## 🎯 Available Presets

| Preset | Resolution | Codec | Bitrate | Use Case |
|--------|------------|-------|---------|----------|
| `720p_av1` | 720p | AV1 | 2Mbps | Excellent compression, smaller files¹ |
| `1080p_av1` | 1080p | AV1 | 4Mbps | Excellent compression, balanced quality¹ |
| `720p_h264` | 720p | H.264 | 3Mbps | Maximum compatibility² |
| `1080p_h264` | 1080p | H.264 | 5Mbps | Maximum compatibility, standard quality² |
| `1080p_h265` | 1080p | H.265 | 3Mbps | Balanced compression and compatibility² |
| `4k_av1` | 4K | AV1 | 15Mbps | Excellent compression for 4K content¹ |
| `4k_h265` | 4K | H.265 | 20Mbps | Balanced compression for 4K content² |

**Notes:**
1. **AV1 encoding**: 
   - NVIDIA systems: Uses `av1_nvenc` hardware acceleration
   - Apple Silicon: Uses optimized `libsvtav1` software encoding
2. **H.264/H.265 encoding**:
   - NVIDIA systems: Uses `h264_nvenc`/`hevc_nvenc` hardware acceleration  
   - Apple Silicon: Uses `h264_videotoolbox`/`hevc_videotoolbox` hardware acceleration

## 🚀 Quick Start

### Basic Usage
```bash
# Transcode a single file
./ffmcli -i movie.mp4 -p 1080p_av1 -o output/

# Process a directory recursively
./ffmcli -i ./videos/ -r -p 720p_h264 -o ./encoded/

# Use multiple workers for faster processing
./ffmcli -i ./videos/ -r -p 1080p_h265 -o ./encoded/ -w 4
```

### Available Commands
```bash
# List all available presets
./ffmcli presets

# Check system capabilities
./ffmcli check

# Show version information
./ffmcli version

# Show help
./ffmcli --help
```

### Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-i, --input` | Input file or directory (required) | - |
| `-o, --output` | Output directory (required) | - |
| `-p, --preset` | Encoding preset | `1080p_h264` |
| `-r, --recursive` | Process directories recursively | `false` |
| `--gpu` | GPU index for multi-GPU systems | `0` |
| `-v, --verbose` | Enable verbose output | `false` |
| `--dry-run` | Preview what would be processed | `false` |
| `--overwrite` | Overwrite existing files | `false` |

## 📖 Examples

### Advanced Usage Examples

```bash
# Dry run to preview processing
./ffmcli -i ./videos/ -r -p 1080p_av1 --dry-run

# Verbose output with specific GPU (NVIDIA systems)
./ffmcli -i video.mp4 -p 4k_h265 -o output/ --gpu 1 -v

# Process with overwrite enabled
./ffmcli -i ./source/ -r -p 720p_av1 -o ./output/ --overwrite

# Maximum parallel processing
./ffmcli -i ./videos/ -r -p 1080p_h264 -o ./encoded/ -w 8

# Force software encoding (disable hardware acceleration)
./ffmcli -i video.mp4 -p 1080p_h264 -o output/ --no-gpu
```

### Apple Silicon Specific Examples

```bash
# Check VideoToolbox availability
./ffmcli check

# Optimal H.264 encoding for Apple Silicon (uses VideoToolbox)
./ffmcli -i input.mp4 -p 1080p_h264 -o output/

# Optimal H.265 encoding for Apple Silicon (uses VideoToolbox)
./ffmcli -i input.mp4 -p 1080p_h265 -o output/

# AV1 encoding on Apple Silicon (uses optimized SVT-AV1)
./ffmcli -i input.mp4 -p 1080p_av1 -o output/

# Batch processing with Apple Silicon optimizations
./ffmcli -i ./videos/ -r -p 1080p_h265 -o ./encoded/ -w 4 -v
```

## 🏗️ Building

### Prerequisites for Building
- Go 1.24 or later
- Make (optional, for using Makefile)

### Build Commands
```bash
# Simple build
go build -o ffmcli main.go

# Using Makefile
make build          # Build the application
make test           # Run tests
make build-all      # Build for multiple platforms
make clean          # Clean build artifacts
```

## 📄 License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments
- Inspired by [Shutter Encoder](https://www.shutterencoder.com/)
- Built with [Cobra CLI](https://github.com/spf13/cobra)
- Utilizes [FFmpeg](https://ffmpeg.org/) for video processing
- **Hardware Acceleration**:
  - NVIDIA NVENC for GPU acceleration
  - Apple VideoToolbox for Apple Silicon acceleration
  - SVT-AV1 for optimized AV1 encoding

**Apple Silicon Performance Notes:**
- H.264/H.265 encoding leverages dedicated media engines for excellent power efficiency
- AV1 encoding uses highly optimized SVT-AV1 software encoder
- M-series chips provide exceptional performance-per-watt for video encoding

## 🔮 Roadmap
Future enhancements being considered:
- [ ] Configuration file support (YAML/JSON)
- [ ] Progress bars and real-time stats
- [ ] Resume interrupted encodings

---
