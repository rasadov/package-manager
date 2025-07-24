package controller

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rasadov/package-manager/config"
	"github.com/rasadov/package-manager/internal/ssh"
	"github.com/rasadov/package-manager/internal/utils"
)

// Create creates a package from the packet configuration
func Create(packetPath string, sshConfig config.SSHConfig) error {
	// Load packet configuration
	packetConfig, err := config.LoadPacketConfig(packetPath)
	if err != nil {
		return fmt.Errorf("failed to load packet config: %w", err)
	}

	fmt.Printf("Creating package: %s (version %s)\n", packetConfig.Name, packetConfig.Version)

	// Collect include and exclude patterns from all targets
	var allIncludePatterns []string
	var allExcludePatterns []string

	for _, target := range packetConfig.Targets {
		allIncludePatterns = append(allIncludePatterns, target.Path)
		allExcludePatterns = append(allExcludePatterns, target.Exclude...)
	}

	if len(allIncludePatterns) == 0 {
		return fmt.Errorf("no targets specified in configuration")
	}

	// Create temporary directory for archive
	tempDir, err := os.MkdirTemp("", "pm-create-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create archive name
	archiveName := fmt.Sprintf("%s-%s.tar.gz", packetConfig.Name, packetConfig.Version)
	archivePath := filepath.Join(tempDir, archiveName)

	// Use your updated CreateTarGz function with include and exclude patterns
	fmt.Printf("Creating archive: %s\n", archiveName)
	fmt.Printf("  Include patterns: %v\n", allIncludePatterns)
	if len(allExcludePatterns) > 0 {
		fmt.Printf("  Exclude patterns: %v\n", allExcludePatterns)
	}

	if err := utils.CreateTarGz(allIncludePatterns, allExcludePatterns, archivePath); err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	// Connect to SSH server
	sshClient := ssh.NewClient(sshConfig)
	if err := sshClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect to SSH server: %w", err)
	}
	defer sshClient.Close()

	// Ensure remote directory exists
	remoteDir := sshClient.GetRemoteDir()
	if err := sshClient.EnsureRemoteDir(remoteDir); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Upload archive
	remotePath := filepath.Join(remoteDir, archiveName)
	fmt.Printf("Uploading to %s...\n", remotePath)

	if err := sshClient.UploadFile(archivePath, remotePath); err != nil {
		return fmt.Errorf("failed to upload archive: %w", err)
	}

	fmt.Printf("Package %s successfully created and uploaded!\n", packetConfig.Name)
	return nil
}
