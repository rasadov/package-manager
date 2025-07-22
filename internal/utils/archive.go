package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

// CreateTarGz creates a tar.gz archive from files matching the given patterns
func CreateTarGz(patterns []string, outputPath string) error {
	files, err := collectFilesByPatterns(patterns)
	if err != nil {
		return fmt.Errorf("failed to collect files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found matching patterns: %v", patterns)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	for _, filePath := range files {
		if err := addFileToTar(tarWriter, filePath); err != nil {
			return fmt.Errorf("failed to add file %s to archive: %w", filePath, err)
		}
	}

	return nil
}

// ExtractTarGz extracts a tar.gz archive to the specified directory
func ExtractTarGz(archivePath, outputDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		if err := extractFileFromTar(tarReader, header, outputDir); err != nil {
			return fmt.Errorf("failed to extract file %s: %w", header.Name, err)
		}
	}

	return nil
}
