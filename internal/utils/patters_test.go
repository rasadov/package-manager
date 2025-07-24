package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestCollectFilesByPatternsWithExclude(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory for consistent testing
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create test files
	testFiles := map[string]string{
		"main.go":            "package main",
		"helper.go":          "package helper",
		"temp.tmp":           "temporary",
		"debug.log":          "log content",
		"config.yaml":        "config: value",
		"internal/cmd.go":    "package internal",
		"internal/temp.tmp":  "internal temp",
		"internal/debug.log": "internal log",
	}

	// Create directories and files
	for path, content := range testFiles {
		dir := filepath.Dir(path)
		if dir != "." {
			os.MkdirAll(dir, 0755)
		}
		os.WriteFile(path, []byte(content), 0644)
	}

	tests := []struct {
		name            string
		includePatterns []string
		excludePatterns []string
		expected        []string
	}{
		{
			name:            "include go files, exclude nothing",
			includePatterns: []string{"*.go"},
			excludePatterns: []string{},
			expected:        []string{"main.go", "helper.go"},
		},
		{
			name:            "include all, exclude tmp files",
			includePatterns: []string{"*"},
			excludePatterns: []string{"*.tmp"},
			expected:        []string{"main.go", "helper.go", "debug.log", "config.yaml"},
		},
		{
			name:            "include all, exclude tmp and log files",
			includePatterns: []string{"*"},
			excludePatterns: []string{"*.tmp", "*.log"},
			expected:        []string{"main.go", "helper.go", "config.yaml"},
		},
		{
			name:            "recursive include with excludes",
			includePatterns: []string{"**/*.go"},
			excludePatterns: []string{"*.tmp", "*.log"},
			expected:        []string{"main.go", "helper.go", "internal/cmd.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := collectFilesByPatternsWithExclude(tt.includePatterns, tt.excludePatterns)
			if err != nil {
				t.Fatalf("collectFilesByPatternsWithExclude() error = %v", err)
			}

			// Convert to relative paths and sort for comparison
			var gotRel []string
			for _, path := range got {
				rel, _ := filepath.Rel(tempDir, path)
				gotRel = append(gotRel, rel)
			}
			sort.Strings(gotRel)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(gotRel, tt.expected) {
				t.Errorf("collectFilesByPatternsWithExclude() = %v, want %v", gotRel, tt.expected)
			}
		})
	}
}

func TestExpandRecursivePattern(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "pm-recursive-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create nested structure
	os.MkdirAll("internal/utils", 0755)
	os.MkdirAll("cmd/pm", 0755)

	files := map[string]string{
		"main.go":                "package main",
		"internal/utils/file.go": "package utils",
		"internal/config.go":     "package internal",
		"cmd/pm/main.go":         "package main",
	}

	for path, content := range files {
		os.WriteFile(path, []byte(content), 0644)
	}

	tests := []struct {
		name     string
		pattern  string
		expected []string
	}{
		{
			name:     "recursive go files",
			pattern:  "**/*.go",
			expected: []string{"main.go", "internal/config.go", "internal/utils/file.go", "cmd/pm/main.go"},
		},
		{
			name:     "recursive in specific directory",
			pattern:  "internal/**/*.go",
			expected: []string{"internal/config.go", "internal/utils/file.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandRecursivePattern(tt.pattern)
			if err != nil {
				t.Fatalf("expandRecursivePattern() error = %v", err)
			}

			sort.Strings(got)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("expandRecursivePattern() = %v, want %v", got, tt.expected)
			}
		})
	}
}
