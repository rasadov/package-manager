package controller

import (
	"fmt"

	"github.com/rasadov/package-manager/config"
	"github.com/rasadov/package-manager/internal/ssh"
)

// Update downloads and installs packages based on packages configuration
func Update(packagesPath string, sshConfig config.SSHConfig) error {
	// Load packages configuration
	packagesConfig, err := config.LoadPackagesConfig(packagesPath)
	if err != nil {
		return fmt.Errorf("failed to load packages config: %w", err)
	}

	fmt.Printf("Updating %d packages...\n", len(packagesConfig.Packages))

	// Connect to SSH server
	sshClient := ssh.NewClient(sshConfig)
	if err := sshClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect to SSH server: %w", err)
	}
	defer sshClient.Close()

	// Process each package
	for _, pkg := range packagesConfig.Packages {
		fmt.Printf("Processing package: %s\n", pkg.Name)

		if err := downloadAndInstallPackage(sshClient, pkg); err != nil {
			fmt.Printf("Warning: Failed to install package %s: %v\n", pkg.Name, err)
			continue
		}

		fmt.Printf("Package %s installed successfully\n", pkg.Name)
	}

	fmt.Println("Package update completed!")
	return nil
}
