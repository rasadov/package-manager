package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// collectFilesByPatterns collects all files matching the given glob patterns (INCLUDE patterns)
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

// collectFilesByPatternsWithExclude collects files matching include patterns but excludes files matching exclude patterns
func collectFilesByPatternsWithExclude(includePatterns []string, excludePatterns []string) ([]string, error) {
	// First collect all files matching include patterns
	allFiles, err := collectFilesByPatterns(includePatterns)
	if err != nil {
		return nil, err
	}

	// If no exclude patterns, return all files
	if len(excludePatterns) == 0 {
		return allFiles, nil
	}

	// Filter out files matching exclude patterns
	var filteredFiles []string
	for _, file := range allFiles {
		shouldExclude := false
		fileName := filepath.Base(file)

		// Check if file matches any exclude pattern
		for _, excludePattern := range excludePatterns {
			matched, err := filepath.Match(excludePattern, fileName)
			if err != nil {
				return nil, fmt.Errorf("invalid exclude pattern %s: %w", excludePattern, err)
			}
			if matched {
				shouldExclude = true
				break
			}
		}

		if !shouldExclude {
			filteredFiles = append(filteredFiles, file)
		}
	}

	return filteredFiles, nil
}
