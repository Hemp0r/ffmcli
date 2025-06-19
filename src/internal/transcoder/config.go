package transcoder

// Config holds the transcoder configuration
type Config struct {
	InputPath      string // Path to input file or directory
	OutputDir      string // Output directory for transcoded files
	Preset         string // Encoding preset name
	GPUIndex       int    // GPU index to use (0-based)
	AudioCodec     string // Audio codec ("copy", "aac", etc.)
	Verbose        bool   // Enable verbose output
	Recursive      bool   // Process files recursively
	Overwrite      bool   // Overwrite existing output files
	NoGPU          bool   // Disable GPU acceleration
	DryRun         bool   // Perform a dry run without actual transcoding
	SkipValidation bool   // Skip path validation (for system checks)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.SkipValidation {
		// Skip validation for system checks
		if c.GPUIndex < 0 {
			c.GPUIndex = 0
		}
		if c.AudioCodec == "" {
			c.AudioCodec = "copy"
		}
		return nil
	}

	if c.InputPath == "" {
		return NewTranscoderError(ErrorTypeInvalidFilePath, "input path is required", nil)
	}
	if c.OutputDir == "" {
		return NewTranscoderError(ErrorTypeInvalidFilePath, "output directory is required", nil)
	}
	if c.GPUIndex < 0 {
		c.GPUIndex = 0
	}
	if c.AudioCodec == "" {
		c.AudioCodec = "copy"
	}
	return nil
}
