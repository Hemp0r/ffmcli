# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.1.0] - 2025-10-16

### Added
- **Delete Source Flag**: New `--delete-source` (`-d`) flag to automatically remove source files after successful transcoding
  - Only deletes files after confirmed successful transcoding
  - Respects dry-run mode (won't delete during `--dry-run`)
  - Provides warning messages if deletion fails
  - Includes verbose logging when enabled
  - Updated CLI help and examples
  - Updated README.md documentation

### Features
- Source file cleanup automation for batch processing workflows
- Disk space management during large transcoding operations
- Safe deletion with comprehensive error handling

### Technical Details
- Added `DeleteSource` field to `Config` struct
- Implemented deletion logic in `processFile` method
- Added CLI flag with both long and short forms
- Integrated with existing verbose logging system

## [v1.0.0] - 2025-10-16

### Added
- Initial release of ffmcli - hardware-accelerated video transcoding tool
- **Multi-Platform Hardware Acceleration**:
  - NVIDIA GPUs with NVENC support (H.264, H.265, AV1)
  - Apple Silicon with VideoToolbox support (H.264, H.265)
  - Optimized software fallback encoding
- **Preset System**: 7 optimized encoding presets:
  - `720p_av1`, `1080p_av1`, `4k_av1` for excellent compression
  - `720p_h264`, `1080p_h264` for maximum compatibility
  - `1080p_h265`, `4k_h265` for balanced compression
- **Processing Features**:
  - Recursive directory processing (`-r/--recursive`)
  - Dry-run mode for preview (`--dry-run`)
  - Verbose output (`-v/--verbose`)
  - Overwrite protection with override (`--overwrite`)
  - Custom audio codec selection (`--audio-codec`)
  - CSV analytics output (`--csv-output`)
- **System Commands**:
  - `check` - System requirements and hardware detection
  - `presets` - List available encoding presets
- **Cross-Platform Support**:
  - Linux, Windows, macOS compatibility
  - Platform-specific optimizations
  - Automatic hardware detection and fallback
- **Error Handling**:
  - Comprehensive error reporting
  - Hardware encoding fallback to software
  - Safe fallback encoding modes
  - Input file validation and probing

### Technical Features
- Built with Go and Cobra CLI framework
- FFmpeg integration with hardware acceleration
- Parallel processing capabilities
- Progress tracking and analytics
- Platform detection and optimization
- Comprehensive test suite