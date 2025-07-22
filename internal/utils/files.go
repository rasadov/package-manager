package utils

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// addFileToTar adds a single file to the tar archive
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

	header.Name = filepath.Base(filePath)

	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}

	if _, err := io.Copy(tarWriter, file); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// extractFileFromTar extracts a single file from tar archive
func extractFileFromTar(tarReader *tar.Reader, header *tar.Header, outputDir string) error {
	targetPath := filepath.Join(outputDir, header.Name)

	if !strings.HasPrefix(targetPath, filepath.Clean(outputDir)+string(os.PathSeparator)) {
		return fmt.Errorf("illegal file path: %s", targetPath)
	}

	switch header.Typeflag {
	case tar.TypeDir:
		if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

	case tar.TypeReg:
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer outFile.Close()

		if _, err := io.Copy(outFile, tarReader); err != nil {
			return fmt.Errorf("failed to write file content: %w", err)
		}

	default:
		return nil
	}

	return nil
}
