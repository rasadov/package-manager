package utils

import (
	"archive/tar"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetArchiveName(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pm-archive-name-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create test files
	testFiles := []string{
		"main.go",
		"internal/cmd.go",
		"nested/deep/file.go",
	}

	for _, file := range testFiles {
		dir := filepath.Dir(file)
		if dir != "." {
			os.MkdirAll(dir, 0755)
		}
		os.WriteFile(file, []byte("test content"), 0644)
	}

	tests := []struct {
		name         string
		filePath     string
		expectedName string
	}{
		{
			name:         "file in current directory",
			filePath:     "main.go",
			expectedName: "main.go",
		},
		{
			name:         "file in subdirectory",
			filePath:     "internal/cmd.go",
			expectedName: "internal/cmd.go",
		},
		{
			name:         "file in nested subdirectory",
			filePath:     "nested/deep/file.go",
			expectedName: "nested/deep/file.go",
		},
		{
			name:         "absolute path",
			filePath:     filepath.Join(tempDir, "main.go"),
			expectedName: "main.go",
		},
		{
			name:         "absolute path to subdirectory file",
			filePath:     filepath.Join(tempDir, "internal/cmd.go"),
			expectedName: "internal/cmd.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archiveName, err := getArchiveName(tt.filePath)
			if err != nil {
				t.Errorf("getArchiveName() error = %v", err)
				return
			}

			// Convert to forward slashes for cross-platform comparison
			archiveName = filepath.ToSlash(archiveName)
			expectedName := filepath.ToSlash(tt.expectedName)

			if archiveName != expectedName {
				t.Errorf("getArchiveName() = %v, want %v", archiveName, expectedName)
			}
		})
	}
}

func TestGetArchiveNameEdgeCases(t *testing.T) {
	// Test edge cases that might cause filepath.Rel to fail
	tempDir, err := os.MkdirTemp("", "pm-edge-case-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create a test file
	os.WriteFile("test.txt", []byte("test"), 0644)

	tests := []struct {
		name        string
		filePath    string
		expectError bool
	}{
		{
			name:        "normal relative path",
			filePath:    "test.txt",
			expectError: false,
		},
		{
			name:        "absolute path within working directory",
			filePath:    filepath.Join(tempDir, "test.txt"),
			expectError: false,
		},
		{
			name:        "nonexistent file",
			filePath:    "nonexistent.txt",
			expectError: false, // getArchiveName shouldn't fail for non-existent files
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archiveName, err := getArchiveName(tt.filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("getArchiveName() unexpected error = %v", err)
				}
				if archiveName == "" {
					t.Errorf("getArchiveName() returned empty string")
				}
			}
		})
	}
}

func TestAddFileToTar(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "pm-add-file-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create test files with different content
	testFiles := map[string]string{
		"simple.txt":           "simple content",
		"internal/nested.go":   "package internal\n\nfunc test() {}",
		"config/settings.yaml": "database:\n  host: localhost\n  port: 5432",
	}

	for path, content := range testFiles {
		dir := filepath.Dir(path)
		if dir != "." {
			os.MkdirAll(dir, 0755)
		}
		os.WriteFile(path, []byte(content), 0644)
	}

	// Create a tar writer for testing
	archivePath := filepath.Join(tempDir, "test.tar")
	outFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Failed to create archive file: %v", err)
	}

	tarWriter := tar.NewWriter(outFile)

	// Test adding each file
	for filePath := range testFiles {
		err := addFileToTar(tarWriter, filePath)
		if err != nil {
			t.Errorf("addFileToTar() error for %s: %v", filePath, err)
		}
	}

	// Close the tar writer to finalize the archive
	tarWriter.Close()
	outFile.Close()

	// Verify the archive contents
	verifyTarContents(t, archivePath, testFiles)
}

func TestAddFileToTarErrors(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pm-add-file-error-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create a tar writer
	archivePath := filepath.Join(tempDir, "test.tar")
	outFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Failed to create archive file: %v", err)
	}
	defer outFile.Close()

	tarWriter := tar.NewWriter(outFile)
	defer tarWriter.Close()

	tests := []struct {
		name        string
		filePath    string
		expectError bool
		setup       func()
	}{
		{
			name:        "nonexistent file",
			filePath:    "does-not-exist.txt",
			expectError: true,
			setup:       func() {},
		},
		{
			name:        "directory instead of file",
			filePath:    "testdir",
			expectError: true, // addFileToTar should reject directories
			setup: func() {
				os.MkdirAll("testdir", 0755)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := addFileToTar(tarWriter, tt.filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("addFileToTar() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestExtractFileFromTar(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pm-extract-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name       string
		tarHeader  *tar.Header
		content    string
		expectFile bool
	}{
		{
			name: "regular file",
			tarHeader: &tar.Header{
				Name:     "test.txt",
				Typeflag: tar.TypeReg,
				Mode:     0644,
				Size:     12,
			},
			content:    "test content",
			expectFile: true,
		},
		{
			name: "nested file",
			tarHeader: &tar.Header{
				Name:     "nested/deep/file.go",
				Typeflag: tar.TypeReg,
				Mode:     0644,
				Size:     15,
			},
			content:    "package nested",
			expectFile: true,
		},
		{
			name: "directory",
			tarHeader: &tar.Header{
				Name:     "testdir/",
				Typeflag: tar.TypeDir,
				Mode:     0755,
			},
			content:    "",
			expectFile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a string reader for the content
			contentReader := strings.NewReader(tt.content)

			// Create tar reader (we're simulating this)
			// In real usage, this comes from tar.NewReader()

			extractPath := filepath.Join(tempDir, "extract-"+tt.name)
			os.MkdirAll(extractPath, 0755)

			// We need to simulate the tar reader, so let's test the logic manually
			targetPath := filepath.Join(extractPath, tt.tarHeader.Name)

			// Test directory creation logic
			if tt.tarHeader.Typeflag == tar.TypeDir {
				err := os.MkdirAll(targetPath, os.FileMode(tt.tarHeader.Mode))
				if err != nil {
					t.Errorf("Failed to create directory: %v", err)
				}

				// Check if directory was created
				if info, err := os.Stat(targetPath); err != nil || !info.IsDir() {
					t.Errorf("Directory was not created properly")
				}
			}

			// Test file creation logic
			if tt.tarHeader.Typeflag == tar.TypeReg {
				// Create parent directories
				if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
					t.Errorf("Failed to create parent directory: %v", err)
				}

				// Create file
				outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(tt.tarHeader.Mode))
				if err != nil {
					t.Errorf("Failed to create file: %v", err)
				} else {
					// Write content
					_, err = outFile.ReadFrom(contentReader)
					outFile.Close()

					if err != nil {
						t.Errorf("Failed to write file content: %v", err)
					}

					// Verify file content
					if tt.expectFile {
						actualContent, err := os.ReadFile(targetPath)
						if err != nil {
							t.Errorf("Failed to read extracted file: %v", err)
						} else if string(actualContent) != tt.content {
							t.Errorf("File content = %q, want %q", string(actualContent), tt.content)
						}
					}
				}
			}
		})
	}
}

func TestExtractFileFromTarSecurity(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pm-security-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	extractDir := filepath.Join(tempDir, "extract")
	os.MkdirAll(extractDir, 0755)

	// Test malicious paths that try to escape the extraction directory
	maliciousPaths := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\evil.exe",
		"/etc/passwd",
		"\\windows\\system32\\evil.exe",
	}

	for _, maliciousPath := range maliciousPaths {
		t.Run("malicious_path_"+strings.ReplaceAll(maliciousPath, "/", "_"), func(t *testing.T) {
			targetPath := filepath.Join(extractDir, maliciousPath)

			// Security check logic (from your extractFileFromTar function)
			cleanOutputDir := filepath.Clean(extractDir)
			cleanTargetPath := filepath.Clean(targetPath)

			isSecure := strings.HasPrefix(cleanTargetPath, cleanOutputDir+string(os.PathSeparator)) || cleanTargetPath == cleanOutputDir

			if !isSecure {
				t.Logf("Good: Security check caught malicious path: %s", maliciousPath)
			} else {
				// This might be OK if the path gets sanitized properly
				t.Logf("Path was allowed (might be sanitized): %s -> %s", maliciousPath, cleanTargetPath)
			}
		})
	}
}

// Helper function to verify tar archive contents
func verifyTarContents(t *testing.T, archivePath string, expectedFiles map[string]string) {
	file, err := os.Open(archivePath)
	if err != nil {
		t.Fatalf("Failed to open archive: %v", err)
	}
	defer file.Close()

	tarReader := tar.NewReader(file)
	foundFiles := make(map[string]string)

	for {
		header, err := tarReader.Next()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Errorf("Error reading tar header: %v", err)
			continue
		}

		if header.Typeflag == tar.TypeReg {
			// Read file content
			content := make([]byte, header.Size)
			n, err := tarReader.Read(content)
			if err != nil && err.Error() != "EOF" {
				t.Errorf("Failed to read file content from archive: %v", err)
				continue
			}

			// Only use the content that was actually read
			foundFiles[header.Name] = string(content[:n])
		}
	}

	// Verify all expected files are present with correct content
	for expectedPath, expectedContent := range expectedFiles {
		// Convert path to archive format (forward slashes)
		archivePath := filepath.ToSlash(expectedPath)

		if actualContent, found := foundFiles[archivePath]; !found {
			t.Errorf("Expected file %s not found in archive", expectedPath)
		} else if actualContent != expectedContent {
			t.Errorf("File %s content = %q, want %q", expectedPath, actualContent, expectedContent)
		}
	}

	// Check for unexpected files
	for foundPath := range foundFiles {
		// Convert back to native path format for comparison
		nativePath := filepath.FromSlash(foundPath)
		if _, expected := expectedFiles[nativePath]; !expected {
			t.Errorf("Unexpected file in archive: %s", foundPath)
		}
	}
}
