package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// collectFilesByPatterns collects all files matching the given glob patterns
func collectFilesByPatterns(patterns []string) ([]string, error) {
	var allFiles []string
	seenFiles := make(map[string]bool)

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %s: %w", pattern, err)
		}

		for _, match := range matches {
			absPath, err := filepath.Abs(match)
			if err != nil {
				return nil, fmt.Errorf("failed to get absolute path for %s: %w", match, err)
			}
			info, err := os.Stat(absPath)
			if err != nil {
				continue
			}
			if info.IsDir() {
				continue
			}

			if !seenFiles[absPath] {
				seenFiles[absPath] = true
				allFiles = append(allFiles, absPath)
			}
		}
	}

	return allFiles, nil
}
