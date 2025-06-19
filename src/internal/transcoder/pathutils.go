package transcoder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathUtils provides utility functions for file paths
type PathUtils struct{}

// NewPathUtils creates a new PathUtils instance
func NewPathUtils() *PathUtils {
	return &PathUtils{}
}

// GenerateOutputPath generates the output file path based on input and preset
func (p *PathUtils) GenerateOutputPath(inputPath, outputDir, inputBasePath string, preset Preset) string {
	filename := filepath.Base(inputPath)
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Sanitize filename - replace problematic characters and limit length
	nameWithoutExt = p.SanitizeFilename(nameWithoutExt)

	var ext string = ".mkv"

	// Create shorter, cleaner filename
	outputFilename := fmt.Sprintf("%s_%s%s", nameWithoutExt, preset.Name, ext)

	// If input is a directory, maintain directory structure
	if info, err := os.Stat(inputBasePath); err == nil && info.IsDir() {
		relPath, err := filepath.Rel(inputBasePath, filepath.Dir(inputPath))
		if err == nil && relPath != "." {
			return filepath.Join(outputDir, relPath, outputFilename)
		}
	}

	return filepath.Join(outputDir, outputFilename)
}

// SanitizeWindowsPath handles long Windows paths and special characters
func (p *PathUtils) SanitizeWindowsPath(path string) string {
	// On Windows, use UNC path for long paths
	if len(path) > 260 && !strings.HasPrefix(path, `\\?\`) {
		// Convert to UNC path for long path support
		absPath, err := filepath.Abs(path)
		if err == nil {
			return `\\?\` + absPath
		}
	}
	return path
}

// SanitizeFilename cleans up filenames for Windows compatibility
func (p *PathUtils) SanitizeFilename(filename string) string {
	problematicChars := map[string]string{
		"<":  "",
		">":  "",
		":":  "",
		"\"": "",
		"|":  "",
		"?":  "",
		"*":  "",
		"/":  "_",
		"\\": "_",
	}

	cleaned := filename
	for old, new := range problematicChars {
		cleaned = strings.ReplaceAll(cleaned, old, new)
	}

	// Replace multiple dots with single underscore to avoid confusion
	cleaned = strings.ReplaceAll(cleaned, "..", "_")
	cleaned = strings.ReplaceAll(cleaned, "...", "_")

	// Limit filename length (Windows has ~255 char limit, leave room for preset and extension)
	if len(cleaned) > 180 {
		cleaned = cleaned[:180]
	}

	// Remove trailing dots and spaces (Windows hates it)
	cleaned = strings.TrimRight(cleaned, ". ")

	return cleaned
}
