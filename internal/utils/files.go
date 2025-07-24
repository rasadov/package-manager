package utils

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// addFileToTar adds a single file to the tar archive while preserving directory structure
func addFileToTar(tarWriter *tar.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return fmt.Errorf("failed to create tar header: %w", err)
	}

	// Get the archive name for this file
	archiveName, err := getArchiveName(filePath)
	if err != nil {
		return fmt.Errorf("failed to get archive name for %s: %w", filePath, err)
	}

	// Use the archive name in the tar header
	header.Name = archiveName

	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}

	if _, err := io.Copy(tarWriter, file); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// getArchiveName determines the name/path to use for a file in the archive
func getArchiveName(filePath string) (string, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Convert both paths to absolute paths for comparison
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for file: %w", err)
	}

	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for current directory: %w", err)
	}

	// Try to get relative path
	relPath, err := filepath.Rel(absCwd, absFilePath)
	if err != nil {
		// If filepath.Rel fails, fall back to a safer approach
		// Check if the file is within the current directory tree
		if strings.HasPrefix(absFilePath, absCwd+string(os.PathSeparator)) {
			// Remove the CWD prefix and leading separator
			relPath = strings.TrimPrefix(absFilePath, absCwd+string(os.PathSeparator))
		} else if absFilePath == absCwd {
			// Edge case: file is exactly the CWD
			relPath = "."
		} else {
			// File is outside the current directory tree
			// Use just the filename as fallback
			relPath = filepath.Base(absFilePath)
		}
	}

	// Ensure we don't have paths that go outside the archive root
	if strings.HasPrefix(relPath, "..") {
		relPath = filepath.Base(absFilePath)
	}

	// Convert to forward slashes for cross-platform compatibility
	archiveName := filepath.ToSlash(relPath)

	// Ensure we don't have empty names
	if archiveName == "" || archiveName == "." {
		archiveName = filepath.Base(absFilePath)
	}

	return archiveName, nil
}

// extractFileFromTar extracts a single file from tar archive while preserving directory structure
func extractFileFromTar(tarReader *tar.Reader, header *tar.Header, outputDir string) error {
	targetPath := filepath.Join(outputDir, header.Name)

	// Security check: ensure the target path is within the output directory
	cleanOutputDir := filepath.Clean(outputDir)
	cleanTargetPath := filepath.Clean(targetPath)
	if !strings.HasPrefix(cleanTargetPath, cleanOutputDir+string(os.PathSeparator)) && cleanTargetPath != cleanOutputDir {
		return fmt.Errorf("illegal file path (directory traversal attempt): %s", targetPath)
	}

	switch header.Typeflag {
	case tar.TypeDir:
		// Create directory
		if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

	case tar.TypeReg:
		// Create parent directories if they don't exist
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		// Create and write the file
		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer outFile.Close()

		if _, err := io.Copy(outFile, tarReader); err != nil {
			return fmt.Errorf("failed to write file content: %w", err)
		}

	default:
		// Skip other file types (symlinks, etc.)
		return nil
	}

	return nil
}
