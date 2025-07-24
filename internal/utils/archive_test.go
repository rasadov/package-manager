package utils

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestCreateTarGz(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pm-archive-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create test files with directory structure
	testFiles := map[string]string{
		"main.go":             "package main\n\nfunc main() {}",
		"config.yaml":         "database:\n  host: localhost",
		"README.md":           "# Test Project",
		"temp.tmp":            "temporary file",
		"debug.log":           "log entry",
		"internal/cmd.go":     "package internal",
		"internal/utils.go":   "package utils",
		"internal/temp.tmp":   "internal temp",
		"nested/deep/file.go": "package deep",
	}

	// Create directories and files
	for path, content := range testFiles {
		dir := filepath.Dir(path)
		if dir != "." {
			os.MkdirAll(dir, 0755)
		}
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", path, err)
		}
	}

	tests := []struct {
		name            string
		includePatterns []string
		excludePatterns []string
		expectedFiles   []string
		expectError     bool
	}{
		{
			name:            "include go files only",
			includePatterns: []string{"*.go"},
			excludePatterns: []string{},
			expectedFiles:   []string{"main.go"},
		},
		{
			name:            "include all go files recursively",
			includePatterns: []string{"**/*.go"},
			excludePatterns: []string{},
			expectedFiles:   []string{"main.go", "internal/cmd.go", "internal/utils.go", "nested/deep/file.go"},
		},
		{
			name:            "include all files, exclude temp files",
			includePatterns: []string{"*"},
			excludePatterns: []string{"*.tmp", "*.log"},
			expectedFiles:   []string{"main.go", "config.yaml", "README.md"},
		},
		{
			name:            "recursive with excludes",
			includePatterns: []string{"**/*"},
			excludePatterns: []string{"*.tmp", "*.log"},
			expectedFiles:   []string{"main.go", "config.yaml", "README.md", "internal/cmd.go", "internal/utils.go", "nested/deep/file.go"},
		},
		{
			name:            "no files match",
			includePatterns: []string{"*.nonexistent"},
			excludePatterns: []string{},
			expectedFiles:   []string{},
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create archive path
			archivePath := filepath.Join(tempDir, "test-archive.tar.gz")

			// Remove archive if it exists from previous test
			os.Remove(archivePath)

			// Create archive
			err := CreateTarGz(tt.includePatterns, tt.excludePatterns, archivePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("CreateTarGz() error = %v", err)
			}

			// Verify archive was created
			if _, err := os.Stat(archivePath); os.IsNotExist(err) {
				t.Fatalf("Archive file was not created")
			}

			// Read archive contents
			actualFiles, err := readTarGzContents(archivePath)
			if err != nil {
				t.Fatalf("Failed to read archive contents: %v", err)
			}

			// Sort for comparison
			sort.Strings(actualFiles)
			sort.Strings(tt.expectedFiles)

			if !reflect.DeepEqual(actualFiles, tt.expectedFiles) {
				t.Errorf("Archive contents = %v, want %v", actualFiles, tt.expectedFiles)
			}

			// Verify file contents by extracting
			extractDir := filepath.Join(tempDir, "extract-test")
			os.MkdirAll(extractDir, 0755)
			defer os.RemoveAll(extractDir)

			err = ExtractTarGz(archivePath, extractDir)
			if err != nil {
				t.Fatalf("Failed to extract archive: %v", err)
			}

			// Verify extracted files match original content
			for _, expectedFile := range tt.expectedFiles {
				extractedPath := filepath.Join(extractDir, expectedFile)
				originalContent := testFiles[expectedFile]

				extractedContent, err := os.ReadFile(extractedPath)
				if err != nil {
					t.Errorf("Failed to read extracted file %s: %v", expectedFile, err)
					continue
				}

				if string(extractedContent) != originalContent {
					t.Errorf("File %s content mismatch. Got %q, want %q",
						expectedFile, string(extractedContent), originalContent)
				}
			}
		})
	}
}

func TestExtractTarGz(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "pm-extract-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test archive manually
	archivePath := filepath.Join(tempDir, "test.tar.gz")
	testFiles := map[string]string{
		"file1.txt":               "content of file 1",
		"subdir/file2.txt":        "content of file 2",
		"subdir/nested/file3.txt": "content of file 3",
	}

	err = createTestArchive(archivePath, testFiles)
	if err != nil {
		t.Fatalf("Failed to create test archive: %v", err)
	}

	// Test extraction
	extractDir := filepath.Join(tempDir, "extracted")
	err = ExtractTarGz(archivePath, extractDir)
	if err != nil {
		t.Fatalf("ExtractTarGz() error = %v", err)
	}

	// Verify extracted files
	for filePath, expectedContent := range testFiles {
		extractedPath := filepath.Join(extractDir, filePath)

		// Check file exists
		if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
			t.Errorf("Extracted file %s does not exist", filePath)
			continue
		}

		// Check content
		content, err := os.ReadFile(extractedPath)
		if err != nil {
			t.Errorf("Failed to read extracted file %s: %v", filePath, err)
			continue
		}

		if string(content) != expectedContent {
			t.Errorf("File %s content = %q, want %q", filePath, string(content), expectedContent)
		}
	}
}

func TestExtractTarGzSecurityCheck(t *testing.T) {
	// Test protection against directory traversal attacks
	tempDir, err := os.MkdirTemp("", "pm-security-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create malicious archive with ../ path traversal
	archivePath := filepath.Join(tempDir, "malicious.tar.gz")
	maliciousFiles := map[string]string{
		"../../../etc/passwd": "malicious content",
		"normal-file.txt":     "normal content",
	}

	err = createTestArchive(archivePath, maliciousFiles)
	if err != nil {
		t.Fatalf("Failed to create malicious archive: %v", err)
	}

	// Try to extract - should fail or sanitize the path
	extractDir := filepath.Join(tempDir, "extracted")
	err = ExtractTarGz(archivePath, extractDir)

	// Should either error or extract safely within extractDir
	if err != nil {
		// If it errors, that's good - it caught the attack
		t.Logf("Good: ExtractTarGz rejected malicious archive: %v", err)
	} else {
		// If it didn't error, check that no files were created outside extractDir
		maliciousPath := filepath.Join(tempDir, "../../../etc/passwd")
		if _, err := os.Stat(maliciousPath); !os.IsNotExist(err) {
			t.Errorf("Security vulnerability: malicious file was created outside extract directory")
		}
	}
}

// Helper function to read tar.gz contents without extracting
func readTarGzContents(archivePath string) ([]string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	var files []string

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Typeflag == tar.TypeReg {
			files = append(files, header.Name)
		}
	}

	return files, nil
}

// Helper function to create a test archive
func createTestArchive(archivePath string, files map[string]string) error {
	outFile, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	for filePath, content := range files {
		header := &tar.Header{
			Name: filePath,
			Mode: 0644,
			Size: int64(len(content)),
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if _, err := tarWriter.Write([]byte(content)); err != nil {
			return err
		}
	}

	return nil
}

func TestArchivePreservesDirectoryStructure(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "pm-structure-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create complex directory structure
	dirs := []string{
		"src/main",
		"src/utils",
		"tests/unit",
		"tests/integration",
		"docs",
	}

	files := map[string]string{
		"src/main/main.go":              "package main",
		"src/main/config.go":            "package main",
		"src/utils/helper.go":           "package utils",
		"tests/unit/main_test.go":       "package main",
		"tests/integration/api_test.go": "package integration",
		"docs/README.md":                "# Documentation",
	}

	// Create structure
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}

	for path, content := range files {
		os.WriteFile(path, []byte(content), 0644)
	}

	// Create archive
	archivePath := filepath.Join(tempDir, "structure-test.tar.gz")
	err = CreateTarGz([]string{"**/*"}, []string{}, archivePath)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}

	// Extract to new location
	extractDir := filepath.Join(tempDir, "extracted")
	err = ExtractTarGz(archivePath, extractDir)
	if err != nil {
		t.Fatalf("Failed to extract archive: %v", err)
	}

	// Verify directory structure is preserved
	for filePath := range files {
		extractedPath := filepath.Join(extractDir, filePath)
		if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
			t.Errorf("File %s not found in extracted archive", filePath)
		}

		// Verify parent directories exist
		dir := filepath.Dir(extractedPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory %s not created during extraction", strings.TrimPrefix(dir, extractDir+"/"))
		}
	}
}
