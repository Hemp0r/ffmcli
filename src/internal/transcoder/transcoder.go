package transcoder

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Transcoder handles video transcoding operations
type Transcoder struct {
	config        Config
	systemChecker *SystemChecker
	fileDiscovery *FileDiscovery
	pathUtils     *PathUtils
	presets       map[string]Preset
}

// New creates a new transcoder instance
func New(config Config) *Transcoder {
	if err := config.Validate(); err != nil {
		// In a real application, you might want to handle this differently
		panic(fmt.Sprintf("Invalid configuration: %v", err))
	}

	executor := &RealCommandExecutor{}
	return &Transcoder{
		config:        config,
		systemChecker: NewSystemChecker(executor),
		fileDiscovery: NewFileDiscovery(),
		pathUtils:     NewPathUtils(),
		presets:       GetPresets(),
	}
}

// CheckFFmpegAvailability checks if FFmpeg is available
func (t *Transcoder) CheckFFmpegAvailability() error {
	return t.systemChecker.CheckFFmpegAvailability()
}

// CheckGPUAvailability checks if NVIDIA GPU is available
func (t *Transcoder) CheckGPUAvailability() error {
	return t.systemChecker.CheckGPUAvailability(t.config.GPUIndex, t.config.Verbose)
}

// CheckEncoderAvailability checks if a specific encoder is available
func (t *Transcoder) CheckEncoderAvailability(encoder string) (bool, error) {
	return t.systemChecker.CheckEncoderAvailability(encoder)
}

// FindVideoFiles finds all video files based on configuration
func (t *Transcoder) FindVideoFiles() ([]string, error) {
	return t.fileDiscovery.FindVideoFiles(t.config.InputPath, t.config.Recursive)
}

// ProcessFiles processes all video files with the configured settings
func (t *Transcoder) ProcessFiles(files []string) error {
	var errors []error

	// Process files sequentially
	for _, file := range files {
		if err := t.processFile(file); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		fmt.Printf("Completed with %d error(s):\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
		return fmt.Errorf("transcoding completed with errors")
	}

	return nil
}

// ProcessFilesWithProgress processes all video files with progress tracking and CSV output
func (t *Transcoder) ProcessFilesWithProgress(files []string, csvWriter *csv.Writer) error {
	total := len(files)
	var errors []error

	// Process files sequentially with progress tracking
	for i, file := range files {
		if err := t.processFileWithAnalytics(file, csvWriter); err != nil {
			errors = append(errors, err)
		}

		// Show progress
		completed := i + 1
		fmt.Printf("Progress: %d/%d files completed (%.1f%%)\n",
			completed, total, float64(completed)/float64(total)*100)
	}

	if len(errors) > 0 {
		fmt.Printf("Completed with %d error(s):\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
		return fmt.Errorf("transcoding completed with errors")
	}

	return nil
}

// processFile processes a single video file
func (t *Transcoder) processFile(inputPath string) error {
	preset, exists := t.presets[t.config.Preset]
	if !exists {
		return NewTranscoderError(ErrorTypeInvalidPreset,
			fmt.Sprintf("preset %s not found", t.config.Preset), nil)
	}

	// Sanitize paths for Windows
	inputPath = t.pathUtils.SanitizeWindowsPath(inputPath)

	// Validate file path for common issues
	if err := ValidateFilePath(inputPath); err != nil {
		return fmt.Errorf("invalid file path: %v", err)
	}

	// Probe input file to ensure it's valid
	if t.config.Verbose {
		fmt.Printf("Probing input file...\n")
	}
	if err := t.probeInputFile(inputPath); err != nil {
		return fmt.Errorf("input file validation failed: %v", err)
	}

	// Generate output filename
	outputPath := t.pathUtils.GenerateOutputPath(inputPath, t.config.OutputDir, t.config.InputPath, preset)
	outputPath = t.pathUtils.SanitizeWindowsPath(outputPath)

	// Check if output already exists
	if !t.config.Overwrite {
		if _, err := os.Stat(outputPath); err == nil {
			if t.config.Verbose {
				fmt.Printf("Skipping %s (output already exists)\n", inputPath)
			}
			return nil
		}
	}

	// Create output directory if needed
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return NewTranscoderError(ErrorTypeFileSystemError,
			"failed to create output directory", err)
	}

	if t.config.Verbose {
		fmt.Printf("Processing: %s -> %s\n", inputPath, outputPath)
	}

	// Build FFmpeg command
	args := t.buildFFmpegArgs(inputPath, outputPath, preset, !t.config.NoGPU)

	// Execute FFmpeg
	startTime := time.Now()
	cmd := exec.Command("ffmpeg", args...)

	if t.config.Verbose {
		encodingMode := "hardware"
		if t.config.NoGPU {
			encodingMode = "software"
		}
		fmt.Printf("Running (%s): ffmpeg %s\n", encodingMode, strings.Join(args, " "))
	}

	// Always capture stderr to get detailed error information
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	ffmpegErr := cmd.Run()
	stderrOutput := stderrBuf.String()

	// Handle encoding errors with fallback
	if ffmpegErr != nil {
		if err := t.handleEncodingError(ffmpegErr, stderrOutput, inputPath, outputPath, preset); err != nil {
			return err
		}
	}

	duration := time.Since(startTime)

	// Get file sizes for compression info
	inputInfo, _ := os.Stat(inputPath)
	outputInfo, _ := os.Stat(outputPath)

	if inputInfo != nil && outputInfo != nil {
		compressionRatio := float64(outputInfo.Size()) / float64(inputInfo.Size()) * 100
		fmt.Printf("Completed %s in %s (%.1f%% of original size)\n",
			filepath.Base(inputPath),
			duration.Round(time.Second),
			compressionRatio)
	}

	return nil
}

// buildFFmpegArgs builds the FFmpeg command arguments
func (t *Transcoder) buildFFmpegArgs(inputPath, outputPath string, preset Preset, useHardware bool) []string {
	args := []string{
		"-hide_banner",
		"-loglevel", "warning",
	}

	platform := t.systemChecker.GetPlatform()

	if useHardware {
		// Add platform-specific hardware acceleration
		switch platform {
		case PlatformAppleSilicon:
			// VideoToolbox doesn't need explicit hwaccel flag, but we can add it for decoding
			args = append(args, "-hwaccel", "videotoolbox")
		case PlatformNVIDIA:
			// Add hardware acceleration for encoding only (avoid hardware decoding issues)
			args = append(args, "-hwaccel", "auto")
		}
	}

	// Add input file
	args = append(args, "-i", inputPath)

	// Add preset arguments (hardware or software)
	if useHardware && (preset.Platform == platform || preset.Platform == Platform(0)) {
		// Use hardware preset if platform matches or preset is platform-agnostic
		args = append(args, preset.Args...)
	} else {
		// Use software encoding
		softwareArgs := t.convertToSoftwarePreset(preset)
		args = append(args, softwareArgs...)
	}

	// Add audio codec
	if t.config.AudioCodec == "" || t.config.AudioCodec == "copy" {
		args = append(args, "-c:a", "copy")
	} else {
		args = append(args, "-c:a", t.config.AudioCodec, "-b:a", "128k")
	}

	// Add output path
	args = append(args, "-y", outputPath)

	return args
}

// handleEncodingError handles FFmpeg encoding errors with fallback strategies
func (t *Transcoder) handleEncodingError(ffmpegErr error, stderrOutput, inputPath, outputPath string, preset Preset) error {
	if !t.config.NoGPU {
		// Try software fallback
		if t.config.Verbose {
			fmt.Printf("Hardware encoding failed, attempting software fallback...\n")
		} else {
			fmt.Printf("Hardware encoding failed for %s, trying software fallback...\n", filepath.Base(inputPath))
		}

		softwareArgs := t.buildFFmpegArgs(inputPath, outputPath, preset, false)
		softwareCmd := exec.Command("ffmpeg", softwareArgs...)

		var softwareStderr strings.Builder
		softwareCmd.Stderr = &softwareStderr

		if err := softwareCmd.Run(); err != nil {
			// Try safe fallback
			safeArgs := t.createSafeFallbackArgs(inputPath, outputPath)
			safeCmd := exec.Command("ffmpeg", safeArgs...)

			if safeErr := safeCmd.Run(); safeErr != nil {
				return NewTranscoderError(ErrorTypeEncodingFailed,
					fmt.Sprintf("all encoding attempts failed for %s", inputPath), safeErr)
			}

			fmt.Printf("Successfully encoded %s using safe fallback mode\n", filepath.Base(inputPath))
		} else {
			fmt.Printf("Successfully encoded %s using software fallback\n", filepath.Base(inputPath))
		}

		return nil
	}

	return NewTranscoderError(ErrorTypeEncodingFailed,
		fmt.Sprintf("encoding failed for %s", inputPath), ffmpegErr)
}

// processFileWithAnalytics processes a single video file and writes analytics to CSV
func (t *Transcoder) processFileWithAnalytics(inputPath string, csvWriter *csv.Writer) error {
	startTime := time.Now()

	// Get input file size
	inputInfo, err := os.Stat(inputPath)
	if err != nil {
		return NewTranscoderError(ErrorTypeFileSystemError,
			"failed to get input file info", err)
	}
	inputSizeMB := float64(inputInfo.Size()) / (1024 * 1024)

	// Process the file using existing method
	err = t.processFile(inputPath)

	endTime := time.Now()
	duration := endTime.Sub(startTime).Seconds()

	// Prepare CSV data
	filename := filepath.Base(inputPath)
	status := "success"
	if err != nil {
		status = "error"
	}

	// Get output file size if successful
	var outputSizeMB float64
	var spaceSavedMB float64
	var compressionRatio float64

	if err == nil {
		preset, exists := t.presets[t.config.Preset]
		if exists {
			outputPath := t.pathUtils.GenerateOutputPath(inputPath, t.config.OutputDir, t.config.InputPath, preset)
			if outputInfo, statErr := os.Stat(outputPath); statErr == nil {
				outputSizeMB = float64(outputInfo.Size()) / (1024 * 1024)
				spaceSavedMB = inputSizeMB - outputSizeMB
				compressionRatio = outputSizeMB / inputSizeMB
			}
		}
	}

	// Write to CSV if provided
	if csvWriter != nil {
		record := []string{
			filename,
			startTime.Format("2006-01-02 15:04:05"),
			endTime.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%.2f", duration),
			fmt.Sprintf("%.2f", inputSizeMB),
			fmt.Sprintf("%.2f", outputSizeMB),
			fmt.Sprintf("%.2f", spaceSavedMB),
			fmt.Sprintf("%.4f", compressionRatio),
			t.config.Preset,
			status,
		}
		if writeErr := csvWriter.Write(record); writeErr != nil {
			fmt.Printf("Warning: failed to write CSV record: %v\n", writeErr)
		}
		csvWriter.Flush()
	}

	return err
}

// convertToSoftwarePreset converts hardware preset arguments to software equivalent
func (t *Transcoder) convertToSoftwarePreset(preset Preset) []string {
	var codec, crf string
	var presetName = "medium"

	switch preset.Encoder {
	// NVIDIA NVENC encoders
	case "h264_nvenc":
		codec = "libx264"
		crf = "23"
	case "hevc_nvenc":
		codec = "libx265"
		crf = "26"
	case "av1_nvenc":
		// Convert to libx264 with higher quality settings (AV1 fallback to H.264)
		codec = "libx264"
		crf = "18"
		presetName = "slower"
	// Apple VideoToolbox encoders
	case "h264_videotoolbox":
		codec = "libx264"
		crf = "23"
	case "hevc_videotoolbox":
		codec = "libx265"
		crf = "26"
	// Software encoders
	case "libsvtav1":
		// SVT-AV1 fallback to libx264
		codec = "libx264"
		crf = "18"
		presetName = "slower"
	default:
		codec = "libx264"
		crf = "23"
	}

	args := []string{
		"-c:v", codec,
		"-preset", presetName,
		"-crf", crf,
		"-vf", t.extractScaleFilter(preset.Args),
	}

	// Add bitrate control if specified
	if preset.Bitrate != "" {
		args = append(args,
			"-b:v", preset.Bitrate,
			"-maxrate", preset.Bitrate,
			"-bufsize", preset.Bitrate,
		)
	}

	return args
}

// extractScaleFilter extracts the scale filter from preset arguments
func (t *Transcoder) extractScaleFilter(args []string) string {
	for i, arg := range args {
		if arg == "-vf" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return "scale=-1:-1" // Default no scaling
}

// probeInputFile probes the input file to check if it's valid and get basic info
func (t *Transcoder) probeInputFile(inputPath string) error {
	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-i", inputPath,
		"-f", "null",
		"-t", "1", // Only check first second
		"-",
	}

	cmd := exec.Command("ffmpeg", args...)
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	if err != nil {
		stderrOutput := stderrBuf.String()
		return NewTranscoderError(ErrorTypeEncodingFailed,
			"input file probe failed", fmt.Errorf("%v\nFFmpeg output: %s", err, stderrOutput))
	}

	return nil
}

// createSafeFallbackArgs creates the simplest possible FFmpeg command that should work
func (t *Transcoder) createSafeFallbackArgs(inputPath, outputPath string) []string {
	return []string{
		"-hide_banner",
		"-loglevel", "error",
		"-i", inputPath,
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "23",
		"-c:a", "copy",
		"-y", outputPath,
	}
}
