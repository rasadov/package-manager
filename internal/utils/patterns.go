package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// collectFilesByPatterns collects all files matching the given glob patterns (INCLUDE patterns)
func collectFilesByPatterns(patterns []string) ([]string, error) {
	var allFiles []string
	seenFiles := make(map[string]bool)

	for _, pattern := range patterns {
		var matches []string
		var err error

		// Check if pattern contains ** for recursive matching
		if strings.Contains(pattern, "**") {
			matches, err = expandRecursivePattern(pattern)
		} else {
			matches, err = filepath.Glob(pattern)
		}

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

// expandRecursivePattern handles ** patterns by walking the directory tree
func expandRecursivePattern(pattern string) ([]string, error) {
	var matches []string

	// Split pattern into parts around **
	parts := strings.Split(pattern, "**")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid recursive pattern: %s (must contain exactly one **)", pattern)
	}

	basePath := parts[0]
	suffix := parts[1]

	// Remove trailing slash from basePath
	if basePath != "" && strings.HasSuffix(basePath, "/") {
		basePath = strings.TrimSuffix(basePath, "/")
	}

	// Remove leading slash from suffix
	if suffix != "" && strings.HasPrefix(suffix, "/") {
		suffix = strings.TrimPrefix(suffix, "/")
	}

	// If basePath is empty, start from current directory
	if basePath == "" {
		basePath = "."
	}

	// Check if basePath exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		// If the base path doesn't exist, return empty results (not an error)
		return matches, nil
	}

	// Walk the directory tree
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip files/directories we can't access
			return nil
		}

		if info.IsDir() {
			// Skip directories - we only want files
			return nil
		}

		// Check if the file matches the suffix pattern
		if suffix == "" {
			// No suffix pattern means match all files
			matches = append(matches, path)
		} else {
			// Extract just the filename for pattern matching
			fileName := filepath.Base(path)

			// Use filepath.Match to check if filename matches the pattern
			matched, err := filepath.Match(suffix, fileName)
			if err != nil {
				// Invalid pattern - skip this file
				return nil
			}

			if matched {
				matches = append(matches, path)
			}
		}

		return nil
	})

	return matches, err
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
