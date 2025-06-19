package transcoder

import (
	"testing"
)

// MockCommandExecutor for testing
type MockCommandExecutor struct {
	shouldFail bool
	output     string
}

func (m *MockCommandExecutor) Execute(name string, args ...string) ([]byte, error) {
	if m.shouldFail {
		return nil, NewTranscoderError(ErrorTypeFFmpegNotFound, "mock error", nil)
	}
	return []byte(m.output), nil
}

func (m *MockCommandExecutor) Run(name string, args ...string) error {
	if m.shouldFail {
		return NewTranscoderError(ErrorTypeFFmpegNotFound, "mock error", nil)
	}
	return nil
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				InputPath: "/test/input",
				OutputDir: "/test/output",
				Preset:    "1080p_h264",
				GPUIndex:  0,
			},
			wantErr: false,
		},
		{
			name: "missing input path",
			config: Config{
				OutputDir: "/test/output",
			},
			wantErr: true,
		},
		{
			name: "missing output dir",
			config: Config{
				InputPath: "/test/input",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSystemChecker_CheckFFmpegAvailability(t *testing.T) {
	tests := []struct {
		name       string
		shouldFail bool
		wantErr    bool
	}{
		{
			name:       "ffmpeg available",
			shouldFail: false,
			wantErr:    false,
		},
		{
			name:       "ffmpeg not available",
			shouldFail: true,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &MockCommandExecutor{shouldFail: tt.shouldFail}
			checker := NewSystemChecker(mockExecutor)

			err := checker.CheckFFmpegAvailability()
			if (err != nil) != tt.wantErr {
				t.Errorf("SystemChecker.CheckFFmpegAvailability() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSystemChecker_CheckGPUAvailability(t *testing.T) {
	tests := []struct {
		name       string
		shouldFail bool
		output     string
		gpuIndex   int
		wantErr    bool
	}{
		{
			name:       "gpu available",
			shouldFail: false,
			output:     "GPU 0: NVIDIA GeForce RTX 3080\n",
			gpuIndex:   0,
			wantErr:    false,
		},
		{
			name:       "gpu not available",
			shouldFail: true,
			output:     "",
			gpuIndex:   0,
			wantErr:    true,
		},
		{
			name:       "gpu index out of range",
			shouldFail: false,
			output:     "GPU 0: NVIDIA GeForce RTX 3080\n",
			gpuIndex:   5,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &MockCommandExecutor{
				shouldFail: tt.shouldFail,
				output:     tt.output,
			}
			checker := NewSystemChecker(mockExecutor)

			err := checker.CheckGPUAvailability(tt.gpuIndex, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("SystemChecker.CheckGPUAvailability() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidPreset(t *testing.T) {
	tests := []struct {
		name   string
		preset string
		want   bool
	}{
		{
			name:   "valid preset",
			preset: "1080p_h264",
			want:   true,
		},
		{
			name:   "invalid preset",
			preset: "invalid_preset",
			want:   false,
		},
		{
			name:   "empty preset",
			preset: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidPreset(tt.preset); got != tt.want {
				t.Errorf("IsValidPreset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAvailablePresets(t *testing.T) {
	presets := GetAvailablePresets()

	if len(presets) == 0 {
		t.Error("GetAvailablePresets() returned empty list")
	}

	// Check that known presets are in the list
	expectedPresets := []string{"1080p_h264", "720p_h264", "1080p_h265", "1080p_av1", "720p_av1", "4k_av1", "4k_h265"}
	for _, expected := range expectedPresets {
		found := false
		for _, preset := range presets {
			if preset == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected preset %s not found in available presets", expected)
		}
	}
}

func TestTranscoderError(t *testing.T) {
	err := NewTranscoderError(ErrorTypeFFmpegNotFound, "test message", nil)

	if err.Type != ErrorTypeFFmpegNotFound {
		t.Errorf("Expected error type %v, got %v", ErrorTypeFFmpegNotFound, err.Type)
	}

	if err.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", err.Message)
	}

	expectedErrorString := "ffmpeg_not_found: test message"
	if err.Error() != expectedErrorString {
		t.Errorf("Expected error string '%s', got '%s'", expectedErrorString, err.Error())
	}
}

func TestPathUtils_SanitizeFilename(t *testing.T) {
	pathUtils := NewPathUtils()

	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "clean filename",
			filename: "video.mp4",
			expected: "video.mp4",
		},
		{
			name:     "filename with problematic chars",
			filename: "video<test>file.mp4",
			expected: "videotestfile.mp4",
		},
		{
			name:     "filename with multiple dots",
			filename: "video...test.mp4",
			expected: "video_.test.mp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pathUtils.SanitizeFilename(tt.filename)
			if result != tt.expected {
				t.Errorf("SanitizeFilename() = %v, want %v", result, tt.expected)
			}
		})
	}
}
