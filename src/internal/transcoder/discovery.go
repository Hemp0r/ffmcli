package transcoder

import (
	"os"
	"path/filepath"
	"strings"
)

// FileDiscovery handles finding video files
type FileDiscovery struct {
	videoExtensions map[string]bool
}

// NewFileDiscovery creates a new file discovery instance
func NewFileDiscovery() *FileDiscovery {
	return &FileDiscovery{
		videoExtensions: map[string]bool{
			".mp4":  true,
			".mkv":  true,
			".avi":  true,
			".mov":  true,
			".wmv":  true,
			".flv":  true,
			".webm": true,
			".m4v":  true,
			".3gp":  true,
			".ts":   true,
			".mts":  true,
			".m2ts": true,
		},
	}
}

// FindVideoFiles finds all video files based on configuration
func (f *FileDiscovery) FindVideoFiles(inputPath string, recursive bool) ([]string, error) {
	var files []string

	info, err := os.Stat(inputPath)
	if err != nil {
		return nil, NewTranscoderError(ErrorTypeFileSystemError,
			"cannot access input path", err)
	}

	if info.IsDir() {
		if recursive {
			err = filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && f.isVideoFile(path) {
					files = append(files, path)
				}
				return nil
			})
		} else {
			entries, err := os.ReadDir(inputPath)
			if err != nil {
				return nil, NewTranscoderError(ErrorTypeFileSystemError,
					"cannot read directory", err)
			}
			for _, entry := range entries {
				if !entry.IsDir() {
					fullPath := filepath.Join(inputPath, entry.Name())
					if f.isVideoFile(fullPath) {
						files = append(files, fullPath)
					}
				}
			}
		}
	} else {
		if f.isVideoFile(inputPath) {
			files = append(files, inputPath)
		}
	}

	return files, err
}

// isVideoFile checks if a file is a video file based on extension
func (f *FileDiscovery) isVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return f.videoExtensions[ext]
}

// ValidateFilePath checks for potential issues with file paths
func ValidateFilePath(path string) error {
	// Check for extremely long paths
	if len(path) > 260 {
		// This is a warning, not an error
		return nil
	}

	// Check for problematic characters in filenames
	filename := filepath.Base(path)
	problematicChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range problematicChars {
		if strings.Contains(filename, char) {
			return NewTranscoderError(ErrorTypeInvalidFilePath,
				"filename contains problematic character", nil)
		}
	}

	return nil
}
