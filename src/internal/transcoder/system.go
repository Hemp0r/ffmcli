package transcoder

import (
	"os/exec"
	"runtime"
	"strings"
)

// Platform represents the hardware platform type
type Platform int

const (
	PlatformUnknown      Platform = iota
	PlatformNVIDIA                // NVIDIA GPU systems
	PlatformAppleSilicon          // Apple Silicon Macs
	PlatformSoftware              // Software-only fallback
)

// CommandExecutor defines an interface for executing external commands
type CommandExecutor interface {
	Execute(name string, args ...string) ([]byte, error)
	Run(name string, args ...string) error
}

// RealCommandExecutor implements CommandExecutor using os/exec
type RealCommandExecutor struct{}

func (r *RealCommandExecutor) Execute(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

func (r *RealCommandExecutor) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// SystemChecker handles system dependency checks
type SystemChecker struct {
	executor CommandExecutor
	platform Platform
}

// NewSystemChecker creates a new system checker
func NewSystemChecker(executor CommandExecutor) *SystemChecker {
	return &SystemChecker{
		executor: executor,
		platform: detectPlatform(),
	}
}

// detectPlatform detects the current hardware platform
func detectPlatform() Platform {
	// Check if we're on macOS with Apple Silicon
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		return PlatformAppleSilicon
	}

	// For other platforms, we'll check for NVIDIA GPU availability
	// This will be determined dynamically in CheckGPUAvailability
	return PlatformUnknown
}

// GetPlatform returns the detected platform
func (s *SystemChecker) GetPlatform() Platform {
	return s.platform
}

// CheckFFmpegAvailability checks if FFmpeg is available
func (s *SystemChecker) CheckFFmpegAvailability() error {
	if err := s.executor.Run("ffmpeg", "-version"); err != nil {
		if s.platform == PlatformAppleSilicon {
			return NewTranscoderError(ErrorTypeFFmpegNotFound,
				"FFmpeg not found. Please install FFmpeg with VideoToolbox support (brew install ffmpeg)", err)
		}
		return NewTranscoderError(ErrorTypeFFmpegNotFound,
			"FFmpeg not found. Please install FFmpeg with NVIDIA support", err)
	}
	return nil
}

// CheckGPUAvailability checks hardware acceleration availability based on platform
func (s *SystemChecker) CheckGPUAvailability(gpuIndex int, verbose bool) error {
	switch s.platform {
	case PlatformAppleSilicon:
		return s.checkAppleSiliconAvailability(verbose)
	default:
		return s.checkNVIDIAAvailability(gpuIndex, verbose)
	}
}

// checkAppleSiliconAvailability checks if VideoToolbox hardware acceleration is available
func (s *SystemChecker) checkAppleSiliconAvailability(verbose bool) error {
	// Check if VideoToolbox encoders are available
	encoders := []string{"h264_videotoolbox", "hevc_videotoolbox"}

	for _, encoder := range encoders {
		available, err := s.CheckEncoderAvailability(encoder)
		if err != nil {
			return NewTranscoderError(ErrorTypeGPUNotAvailable,
				"Failed to check VideoToolbox encoder availability", err)
		}
		if !available {
			return NewTranscoderError(ErrorTypeGPUNotAvailable,
				"VideoToolbox encoders not available. Please ensure FFmpeg is built with VideoToolbox support", nil)
		}
	}

	if verbose {
		output, _ := s.executor.Execute("system_profiler", "SPHardwareDataType")
		if strings.Contains(string(output), "Apple") {
			// VideoToolbox is available on Apple Silicon
			return nil
		}
	}

	return nil
}

// checkNVIDIAAvailability checks if NVIDIA GPU is available (original implementation)
func (s *SystemChecker) checkNVIDIAAvailability(gpuIndex int, verbose bool) error {
	output, err := s.executor.Execute("nvidia-smi", "-L")
	if err != nil {
		// Update platform to software fallback if NVIDIA not available
		s.platform = PlatformSoftware
		return NewTranscoderError(ErrorTypeGPUNotAvailable,
			"NVIDIA GPU not detected. Please ensure NVIDIA drivers are installed", err)
	}

	// Parse GPU list
	lines := strings.Split(string(output), "\n")
	gpuCount := 0
	for _, line := range lines {
		if strings.Contains(line, "GPU") {
			gpuCount++
		}
	}

	if gpuCount == 0 {
		s.platform = PlatformSoftware
		return NewTranscoderError(ErrorTypeGPUNotAvailable, "no NVIDIA GPUs found", nil)
	}

	if gpuIndex >= gpuCount {
		return NewTranscoderError(ErrorTypeGPUNotAvailable,
			"GPU index not available", nil)
	}

	// Update platform to NVIDIA if successful
	s.platform = PlatformNVIDIA
	return nil
}

// CheckEncoderAvailability checks if a specific encoder is available
func (s *SystemChecker) CheckEncoderAvailability(encoder string) (bool, error) {
	output, err := s.executor.Execute("ffmpeg", "-encoders")
	if err != nil {
		return false, NewTranscoderError(ErrorTypeEncoderNotFound,
			"failed to check encoders", err)
	}

	return strings.Contains(string(output), encoder), nil
}
