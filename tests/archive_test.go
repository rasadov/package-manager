package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/rasadov/package-manager/internal/utils"
)

func TestCreateTarGz(t *testing.T) {
	testDir := t.TempDir()

	testFiles := map[string]string{
		"file1.txt":        "content of file 1",
		"file2.go":         "package main\n\nfunc main() {}",
		"config.json":      `{"name": "test"}`,
		"subdir/file3.txt": "content in subdirectory",
		"ignore.tmp":       "this should be ignored",
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(testDir, filePath)

		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", fullPath, err)
		}
	}

	tests := []struct {
		name        string
		patterns    []string
		expectFiles []string
		expectError bool
	}{
		{
			name:        "Single pattern - txt files",
			patterns:    []string{filepath.Join(testDir, "*.txt")},
			expectFiles: []string{"file1.txt"},
			expectError: false,
		},
		{
			name: "Multiple patterns",
			patterns: []string{
				filepath.Join(testDir, "*.txt"),
				filepath.Join(testDir, "*.go"),
				filepath.Join(testDir, "*.json"),
			},
			expectFiles: []string{"file1.txt", "file2.go", "config.json"},
			expectError: false,
		},
		{
			name:        "Pattern with subdirectory",
			patterns:    []string{filepath.Join(testDir, "subdir", "*.txt")},
			expectFiles: []string{"file3.txt"},
			expectError: false,
		},
		{
			name:        "No matching files",
			patterns:    []string{filepath.Join(testDir, "*.nonexistent")},
			expectFiles: nil,
			expectError: true,
		},
		{
			name:        "Invalid pattern",
			patterns:    []string{"[invalid"},
			expectFiles: nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(testDir, tt.name+".tar.gz")

			err := utils.CreateTarGz(tt.patterns, outputPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify archive was created
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Errorf("Archive file was not created: %s", outputPath)
			}

			// Verify archive is not empty
			info, err := os.Stat(outputPath)
			if err != nil {
				t.Fatalf("Failed to stat archive: %v", err)
			}
			if info.Size() == 0 {
				t.Errorf("Archive file is empty")
			}
		})
	}
}

func TestExtractTarGz(t *testing.T) {
	testDir := t.TempDir()

	testFiles := map[string]string{
		"test1.txt": "Hello, World!",
		"test2.go":  "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}",
		"config.json": `{
	"name": "test-package",
	"version": "1.0.0"
}`,
	}

	for fileName, content := range testFiles {
		filePath := filepath.Join(testDir, fileName)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	archivePath := filepath.Join(testDir, "test.tar.gz")
	patterns := []string{filepath.Join(testDir, "*")}

	if err := utils.CreateTarGz(patterns, archivePath); err != nil {
		t.Fatalf("Failed to create test archive: %v", err)
	}

	tests := []struct {
		name        string
		archivePath string
		outputDir   string
		expectError bool
		expectFiles []string
	}{
		{
			name:        "Valid extraction",
			archivePath: archivePath,
			outputDir:   filepath.Join(testDir, "extract1"),
			expectError: false,
			expectFiles: []string{"test1.txt", "test2.go", "config.json"},
		},
		{
			name:        "Extract to different directory",
			archivePath: archivePath,
			outputDir:   filepath.Join(testDir, "extract2"),
			expectError: false,
			expectFiles: []string{"test1.txt", "test2.go", "config.json"},
		},
		{
			name:        "Non-existent archive",
			archivePath: filepath.Join(testDir, "nonexistent.tar.gz"),
			outputDir:   filepath.Join(testDir, "extract3"),
			expectError: true,
			expectFiles: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.MkdirAll(tt.outputDir, 0755); err != nil {
				t.Fatalf("Failed to create output directory: %v", err)
			}

			err := utils.ExtractTarGz(tt.archivePath, tt.outputDir)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify extracted files
			for _, expectedFile := range tt.expectFiles {
				extractedPath := filepath.Join(tt.outputDir, expectedFile)

				// Check file exists
				if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
					t.Errorf("Expected file not found: %s", extractedPath)
					continue
				}

				// Check file content matches original
				originalContent := testFiles[expectedFile]
				extractedContent, err := os.ReadFile(extractedPath)
				if err != nil {
					t.Errorf("Failed to read extracted file %s: %v", extractedPath, err)
					continue
				}

				if string(extractedContent) != originalContent {
					t.Errorf("Content mismatch for file %s.\nExpected: %s\nGot: %s",
						expectedFile, originalContent, string(extractedContent))
				}
			}
		})
	}
}

func TestCreateAndExtractRoundTrip(t *testing.T) {
	testDir := t.TempDir()

	testFiles := map[string]string{
		"main.go":          "package main\n\nfunc main() { println(\"Hello\") }",
		"config.json":      `{"name": "test", "version": "1.0"}`,
		"readme.txt":       "This is a readme file",
		"subdir/nested.go": "package subdir\n\nconst Value = 42",
		"data.csv":         "name,age\nJohn,30\nJane,25",
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(testDir, filePath)

		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Test patterns
	patterns := []string{
		filepath.Join(testDir, "*.go"),
		filepath.Join(testDir, "*.json"),
		filepath.Join(testDir, "*.txt"),
		filepath.Join(testDir, "subdir", "*.go"),
	}

	archivePath := filepath.Join(testDir, "roundtrip.tar.gz")
	if err := utils.CreateTarGz(patterns, archivePath); err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}
	extractDir := filepath.Join(testDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatalf("Failed to create extract directory: %v", err)
	}

	if err := utils.ExtractTarGz(archivePath, extractDir); err != nil {
		t.Fatalf("Failed to extract archive: %v", err)
	}

	// Verify extracted files
	expectedFiles := []string{
		"main.go",
		"config.json",
		"readme.txt",
		"nested.go",
	}

	for _, expectedFile := range expectedFiles {
		extractedPath := filepath.Join(extractDir, expectedFile)

		if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
			t.Errorf("Expected extracted file not found: %s", extractedPath)
			continue
		}
		content, err := os.ReadFile(extractedPath)
		if err != nil {
			t.Errorf("Failed to read extracted file: %v", err)
			continue
		}

		if len(content) == 0 {
			t.Errorf("Extracted file %s is empty", expectedFile)
		}
	}

	// Verify data.csv was NOT included (not in patterns)
	excludedFile := filepath.Join(extractDir, "data.csv")
	if _, err := os.Stat(excludedFile); err == nil {
		t.Errorf("File that should be excluded was found: %s", excludedFile)
	}
}

// Benchmark tests
func BenchmarkCreateTarGz(b *testing.B) {
	testDir := b.TempDir()

	for i := 0; i < 100; i++ {
		fileName := filepath.Join(testDir, fmt.Sprintf("file%d.txt", i))
		content := fmt.Sprintf("This is test file number %d", i)
		if err := os.WriteFile(fileName, []byte(content), 0644); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
	}

	patterns := []string{filepath.Join(testDir, "*.txt")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		archivePath := filepath.Join(testDir, fmt.Sprintf("bench%d.tar.gz", i))
		if err := utils.CreateTarGz(patterns, archivePath); err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
