package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"ffmcli/internal/transcoder"

	"github.com/spf13/cobra"
)

var (
	recursive  bool
	outputDir  string
	preset     string
	inputFile  string
	overwrite  bool
	verbose    bool
	dryRun     bool
	gpuIndex   int
	noGPU      bool
	audioCodec string
	csvOutput  string
)

var rootCmd = &cobra.Command{
	Use:   "ffmcli",
	Short: "A hardware-accelerated video transcoder",
	Long: `ffmcli is a command-line tool for transcoding video files using hardware acceleration.
It supports NVIDIA GPUs (NVENC) and Apple Silicon (VideoToolbox) for fast, efficient encoding.
Includes recursive directory scanning and provides presets for common encoding scenarios.`,
	Example: `  # Transcode a single file to 1080p AV1
  ffmcli -i input.mp4 -p 1080p_av1 -o output/

  # Recursively transcode all videos in a directory
  ffmcli -i /path/to/videos/ -r -p 720p_av1 -o /path/to/output/

  # Dry run to see what would be processed
  ffmcli -i /path/to/videos/ -r -p 1080p_h264 --dry-run

  # Force software encoding (disable GPU)
  ffmcli -i input.mp4 -p 1080p_h264 -o output/ --no-gpu`,
	RunE: runTranscode,
}

func init() {
	rootCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file or directory (required)")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory (required)")
	rootCmd.Flags().StringVarP(&preset, "preset", "p", "1080p_h264", "Encoding preset (720p_av1, 1080p_av1, 720p_h264, 1080p_h264, 1080p_h265, 4k_av1, 4k_h265)")
	rootCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively process directories")
	rootCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing output files")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be processed without actually transcoding")
	rootCmd.Flags().IntVar(&gpuIndex, "gpu", 0, "GPU index to use (default: 0)")
	rootCmd.Flags().BoolVar(&noGPU, "no-gpu", false, "Force software encoding (disable GPU acceleration)")
	rootCmd.Flags().StringVar(&audioCodec, "audio-codec", "copy", "Audio codec: copy (default), aac, ac3, mp3")
	rootCmd.Flags().StringVar(&csvOutput, "csv-output", "", "CSV file to save conversion analytics (optional)")

	rootCmd.MarkFlagRequired("input")
	rootCmd.MarkFlagRequired("output")

	// Add subcommands
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(presetsCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func runTranscode(cmd *cobra.Command, args []string) error {
	// Validate required flags
	if inputFile == "" {
		return fmt.Errorf("input file or directory is required")
	}
	if outputDir == "" {
		return fmt.Errorf("output directory is required")
	}

	// Check if input exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file or directory does not exist: %s", inputFile)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Validate preset
	if !transcoder.IsValidPreset(preset) {
		availablePresets := strings.Join(transcoder.GetAvailablePresets(), ", ")
		return fmt.Errorf("invalid preset '%s'. Available presets: %s", preset, availablePresets)
	}

	// Create transcoder config
	config := transcoder.Config{
		InputPath:  inputFile,
		OutputDir:  outputDir,
		Preset:     preset,
		Recursive:  recursive,
		Overwrite:  overwrite,
		Verbose:    verbose,
		DryRun:     dryRun,
		GPUIndex:   gpuIndex,
		NoGPU:      noGPU,
		AudioCodec: audioCodec,
	}

	// Initialize transcoder
	t := transcoder.New(config)

	// Check GPU availability (skip if using software-only mode)
	if !noGPU {
		if err := t.CheckGPUAvailability(); err != nil {
			fmt.Printf("GPU check failed: %v\n", err)
			fmt.Printf("Consider using --no-gpu flag for software encoding\n")
			return err
		}
	}

	// Find files to process
	files, err := t.FindVideoFiles()
	if err != nil {
		return fmt.Errorf("failed to find video files: %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no video files found")
	}

	fmt.Printf("Found %d video file(s) to process\n", len(files))

	// Setup CSV logging if requested
	var csvWriter *csv.Writer
	var csvFile *os.File
	if csvOutput != "" {
		var err error
		csvFile, err = os.Create(csvOutput)
		if err != nil {
			return fmt.Errorf("failed to create CSV file: %v", err)
		}
		defer csvFile.Close()

		csvWriter = csv.NewWriter(csvFile)
		defer csvWriter.Flush()

		// Write CSV header
		header := []string{"filename", "start_time", "end_time", "duration_seconds", "size_before_mb", "size_after_mb", "space_saved_mb", "compression_ratio", "preset", "status"}
		if err := csvWriter.Write(header); err != nil {
			return fmt.Errorf("failed to write CSV header: %v", err)
		}
	}

	// Process files with progress tracking
	return t.ProcessFilesWithProgress(files, csvWriter)
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check system requirements and hardware acceleration availability",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := transcoder.Config{SkipValidation: true}
		t := transcoder.New(config)

		fmt.Println("System Check Results:")
		fmt.Println("====================")

		// Check FFmpeg
		if err := t.CheckFFmpegAvailability(); err != nil {
			fmt.Printf("FFmpeg: %v\n", err)
		} else {
			fmt.Println("FFmpeg: Available")
		}

		// Check hardware acceleration
		if err := t.CheckGPUAvailability(); err != nil {
			fmt.Printf("Hardware Acceleration: %v\n", err)
		} else {
			// Determine platform and show appropriate message
			systemChecker := transcoder.NewSystemChecker(&transcoder.RealCommandExecutor{})
			platform := systemChecker.GetPlatform()

			switch platform {
			case transcoder.PlatformAppleSilicon:
				fmt.Println("Hardware Acceleration: Apple Silicon VideoToolbox detected")
			case transcoder.PlatformNVIDIA:
				fmt.Println("Hardware Acceleration: NVIDIA GPU with CUDA support detected")
			default:
				fmt.Println("Hardware Acceleration: Available")
			}
		}

		// Check platform-appropriate encoders
		systemChecker := transcoder.NewSystemChecker(&transcoder.RealCommandExecutor{})
		platform := systemChecker.GetPlatform()

		var encoders []string
		switch platform {
		case transcoder.PlatformAppleSilicon:
			encoders = []string{"h264_videotoolbox", "hevc_videotoolbox", "libsvtav1"}
		default:
			encoders = []string{"h264_nvenc", "hevc_nvenc", "av1_nvenc"}
		}

		for _, encoder := range encoders {
			if available, err := t.CheckEncoderAvailability(encoder); err != nil {
				fmt.Printf("%s: Error checking (%v)\n", encoder, err)
			} else if available {
				fmt.Printf("%s: Available\n", encoder)
			} else {
				fmt.Printf("%s: Not available\n", encoder)
			}
		}

		return nil
	},
}

var presetsCmd = &cobra.Command{
	Use:   "presets",
	Short: "List available encoding presets",
	RunE: func(cmd *cobra.Command, args []string) error {
		presets := transcoder.GetAvailablePresets()

		fmt.Println("Available Presets:")
		fmt.Println("==================")

		for _, preset := range presets {
			fmt.Printf("  %s\n", preset)
		}

		fmt.Println("\nExample Usage:")
		fmt.Println("  ffmcli -i input.mp4 -p 1080p_av1 -o output/")

		return nil
	},
}
