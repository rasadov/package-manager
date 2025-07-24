package controller

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rasadov/package-manager/config"
	"github.com/rasadov/package-manager/internal/ssh"
	"github.com/rasadov/package-manager/internal/utils"
)

// downloadAndInstallPackage downloads and extracts a single package
func downloadAndInstallPackage(sshClient *ssh.Client, pkg config.PackageRequest) error {
	// Find the best matching package version on server
	archiveName, err := findBestPackageVersion(sshClient, pkg)
	if err != nil {
		return fmt.Errorf("failed to find package version: %w", err)
	}

	// Create temporary directory for download
	tempDir, err := os.MkdirTemp("", "pm-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download archive
	remotePath := filepath.Join(sshClient.GetRemoteDir(), archiveName)
	localPath := filepath.Join(tempDir, archiveName)

	fmt.Printf("Downloading %s...\n", archiveName)
	if err := sshClient.DownloadFile(remotePath, localPath); err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}

	// Create installation directory
	installDir := filepath.Join("packages", pkg.Name)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// Use your existing ExtractTarGz function
	fmt.Printf("Extracting %s to %s...\n", archiveName, installDir)
	if err := utils.ExtractTarGz(localPath, installDir); err != nil {
		return fmt.Errorf("failed to extract package: %w", err)
	}

	return nil
}

// findBestPackageVersion finds the best matching package version on the server
func findBestPackageVersion(sshClient *ssh.Client, pkg config.PackageRequest) (string, error) {
	// List files in remote directory
	files, err := sshClient.ListFiles(sshClient.GetRemoteDir())
	if err != nil {
		return "", fmt.Errorf("failed to list remote files: %w", err)
	}

	// Find matching packages
	var candidates []string
	prefix := pkg.Name + "-"
	suffix := ".tar.gz"

	for _, file := range files {
		if strings.HasPrefix(file, prefix) && strings.HasSuffix(file, suffix) {
			candidates = append(candidates, file)
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no packages found for %s", pkg.Name)
	}

	// For now, return the first match
	// TODO: Implement proper version comparison based on pkg.Version constraints
	fmt.Printf("Found %d candidate(s) for %s, selecting: %s\n", len(candidates), pkg.Name, candidates[0])
	return candidates[0], nil
}
