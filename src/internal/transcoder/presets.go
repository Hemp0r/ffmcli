package transcoder

type Preset struct {
	Name        string   // Preset name (e.g., "720p_av1")
	Resolution  string   // Target resolution (e.g., "1280x720")
	Codec       string   // Video codec (e.g., "AV1", "H.264", "H.265")
	Encoder     string   // FFmpeg encoder name (e.g., "av1_nvenc")
	Bitrate     string   // Target bitrate (e.g., "2M")
	Description string   // Human-readable description
	Args        []string // FFmpeg command line arguments
	Platform    Platform // Target platform for this preset
}

func GetPresets() map[string]Preset {
	// Detect current platform
	platform := detectPlatform()

	presets := make(map[string]Preset)

	// Add platform-appropriate presets
	switch platform {
	case PlatformAppleSilicon:
		addAppleSiliconPresets(presets)
	default:
		addNVIDIAPresets(presets)
	}

	return presets
}

// addNVIDIAPresets adds NVIDIA NVENC presets (original implementation)
func addNVIDIAPresets(presets map[string]Preset) {
	nvencPresets := map[string]Preset{
		"720p_av1": {
			Name:        "720p_av1",
			Resolution:  "1280x720",
			Codec:       "AV1",
			Encoder:     "av1_nvenc",
			Bitrate:     "2M",
			Description: "720p AV1 encoding with NVENC",
			Args:        []string{"-c:v", "av1_nvenc", "-preset", "p7", "-crf", "28", "-b:v", "2M", "-maxrate", "3M", "-bufsize", "6M", "-vf", "scale=1280:720"},
			Platform:    PlatformNVIDIA,
		},
		"1080p_av1": {
			Name:        "1080p_av1",
			Resolution:  "1920x1080",
			Codec:       "AV1",
			Encoder:     "av1_nvenc",
			Bitrate:     "4M",
			Description: "1080p AV1 encoding with NVENC",
			Args:        []string{"-c:v", "av1_nvenc", "-preset", "p7", "-crf", "26", "-b:v", "4M", "-maxrate", "6M", "-bufsize", "12M", "-vf", "scale=1920:1080"},
			Platform:    PlatformNVIDIA,
		},
		"720p_h264": {
			Name:        "720p_h264",
			Resolution:  "1280x720",
			Codec:       "H.264",
			Encoder:     "h264_nvenc",
			Bitrate:     "3M",
			Description: "720p H.264 encoding with NVENC",
			Args:        []string{"-c:v", "h264_nvenc", "-preset", "p7", "-crf", "23", "-b:v", "3M", "-maxrate", "4M", "-bufsize", "8M", "-vf", "scale=1280:720"},
			Platform:    PlatformNVIDIA,
		},
		"1080p_h264": {
			Name:        "1080p_h264",
			Resolution:  "1920x1080",
			Codec:       "H.264",
			Encoder:     "h264_nvenc",
			Bitrate:     "5M",
			Description: "1080p H.264 encoding with NVENC",
			Args:        []string{"-c:v", "h264_nvenc", "-preset", "p7", "-crf", "23", "-b:v", "5M", "-maxrate", "8M", "-bufsize", "16M", "-vf", "scale=1920:1080"},
			Platform:    PlatformNVIDIA,
		},
		"1080p_h265": {
			Name:        "1080p_h265",
			Resolution:  "1920x1080",
			Codec:       "H.265",
			Encoder:     "hevc_nvenc",
			Bitrate:     "3M",
			Description: "1080p H.265 encoding with NVENC",
			Args:        []string{"-c:v", "hevc_nvenc", "-preset", "p7", "-crf", "26", "-b:v", "3M", "-maxrate", "5M", "-bufsize", "10M", "-vf", "scale=1920:1080"},
			Platform:    PlatformNVIDIA,
		},
		"4k_av1": {
			Name:        "4k_av1",
			Resolution:  "3840x2160",
			Codec:       "AV1",
			Encoder:     "av1_nvenc",
			Bitrate:     "15M",
			Description: "4K AV1 encoding with NVENC",
			Args:        []string{"-c:v", "av1_nvenc", "-preset", "p7", "-crf", "28", "-b:v", "15M", "-maxrate", "20M", "-bufsize", "40M", "-vf", "scale=3840:2160"},
			Platform:    PlatformNVIDIA,
		},
		"4k_h265": {
			Name:        "4k_h265",
			Resolution:  "3840x2160",
			Codec:       "H.265",
			Encoder:     "hevc_nvenc",
			Bitrate:     "20M",
			Description: "4K H.265 encoding with NVENC",
			Args:        []string{"-c:v", "hevc_nvenc", "-preset", "p7", "-crf", "26", "-b:v", "20M", "-maxrate", "30M", "-bufsize", "60M", "-vf", "scale=3840:2160"},
			Platform:    PlatformNVIDIA,
		},
	}

	for name, preset := range nvencPresets {
		presets[name] = preset
	}
}

// addAppleSiliconPresets adds Apple Silicon VideoToolbox presets
func addAppleSiliconPresets(presets map[string]Preset) {
	appleSiliconPresets := map[string]Preset{
		"720p_av1": {
			Name:        "720p_av1",
			Resolution:  "1280x720",
			Codec:       "AV1",
			Encoder:     "libsvtav1", // Use software AV1 encoder (Apple Silicon doesn't have native AV1 VideoToolbox)
			Bitrate:     "2M",
			Description: "720p AV1 encoding (software, optimized for Apple Silicon)",
			Args:        []string{"-c:v", "libsvtav1", "-preset", "6", "-crf", "28", "-b:v", "2M", "-maxrate", "3M", "-bufsize", "6M", "-vf", "scale=1280:720"},
			Platform:    PlatformAppleSilicon,
		},
		"1080p_av1": {
			Name:        "1080p_av1",
			Resolution:  "1920x1080",
			Codec:       "AV1",
			Encoder:     "libsvtav1", // Use software AV1 encoder
			Bitrate:     "4M",
			Description: "1080p AV1 encoding (software, optimized for Apple Silicon)",
			Args:        []string{"-c:v", "libsvtav1", "-preset", "6", "-crf", "26", "-b:v", "4M", "-maxrate", "6M", "-bufsize", "12M", "-vf", "scale=1920:1080"},
			Platform:    PlatformAppleSilicon,
		},
		"720p_h264": {
			Name:        "720p_h264",
			Resolution:  "1280x720",
			Codec:       "H.264",
			Encoder:     "h264_videotoolbox",
			Bitrate:     "3M",
			Description: "720p H.264 encoding with VideoToolbox",
			Args:        []string{"-c:v", "h264_videotoolbox", "-q:v", "65", "-b:v", "3M", "-maxrate", "4M", "-bufsize", "8M", "-vf", "scale=1280:720"},
			Platform:    PlatformAppleSilicon,
		},
		"1080p_h264": {
			Name:        "1080p_h264",
			Resolution:  "1920x1080",
			Codec:       "H.264",
			Encoder:     "h264_videotoolbox",
			Bitrate:     "5M",
			Description: "1080p H.264 encoding with VideoToolbox",
			Args:        []string{"-c:v", "h264_videotoolbox", "-q:v", "65", "-b:v", "5M", "-maxrate", "8M", "-bufsize", "16M", "-vf", "scale=1920:1080"},
			Platform:    PlatformAppleSilicon,
		},
		"1080p_h265": {
			Name:        "1080p_h265",
			Resolution:  "1920x1080",
			Codec:       "H.265",
			Encoder:     "hevc_videotoolbox",
			Bitrate:     "3M",
			Description: "1080p H.265 encoding with VideoToolbox",
			Args:        []string{"-c:v", "hevc_videotoolbox", "-q:v", "65", "-b:v", "3M", "-maxrate", "5M", "-bufsize", "10M", "-vf", "scale=1920:1080"},
			Platform:    PlatformAppleSilicon,
		},
		"4k_av1": {
			Name:        "4k_av1",
			Resolution:  "3840x2160",
			Codec:       "AV1",
			Encoder:     "libsvtav1", // Use software AV1 encoder
			Bitrate:     "15M",
			Description: "4K AV1 encoding (software, optimized for Apple Silicon)",
			Args:        []string{"-c:v", "libsvtav1", "-preset", "5", "-crf", "28", "-b:v", "15M", "-maxrate", "20M", "-bufsize", "40M", "-vf", "scale=3840:2160"},
			Platform:    PlatformAppleSilicon,
		},
		"4k_h265": {
			Name:        "4k_h265",
			Resolution:  "3840x2160",
			Codec:       "H.265",
			Encoder:     "hevc_videotoolbox",
			Bitrate:     "20M",
			Description: "4K H.265 encoding with VideoToolbox",
			Args:        []string{"-c:v", "hevc_videotoolbox", "-q:v", "60", "-b:v", "20M", "-maxrate", "30M", "-bufsize", "60M", "-vf", "scale=3840:2160"},
			Platform:    PlatformAppleSilicon,
		},
	}

	for name, preset := range appleSiliconPresets {
		presets[name] = preset
	}
}

// Global cache to avoid repeated calls to GetPresets()
var presetCache map[string]Preset
var presetNames []string

func init() {
	presetCache = GetPresets()
	presetNames = make([]string, 0, len(presetCache))
	for name := range presetCache {
		presetNames = append(presetNames, name)
	}
}

func IsValidPreset(preset string) bool {
	_, exists := presetCache[preset]
	return exists
}

func GetAvailablePresets() []string {
	return presetNames
}

// GetPresetsForPlatform returns presets suitable for the specified platform
func GetPresetsForPlatform(platform Platform) map[string]Preset {
	allPresets := GetPresets()
	filteredPresets := make(map[string]Preset)

	for name, preset := range allPresets {
		if preset.Platform == platform {
			filteredPresets[name] = preset
		}
	}

	return filteredPresets
}
