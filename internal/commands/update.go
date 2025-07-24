package commands

import (
	"fmt"
	"os"

	"github.com/rasadov/package-manager/config"
	"github.com/rasadov/package-manager/internal/controller"
	"github.com/spf13/cobra"
)

func Update() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "update <packages.json>",
		Short: "Download and install packages",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packagesPath := args[0]

			// Check if packages file exists
			if _, err := os.Stat(packagesPath); os.IsNotExist(err) {
				return fmt.Errorf("packages file not found: %s", packagesPath)
			}

			// Load SSH configuration
			sshConfig, err := config.LoadSSHConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load SSH config: %w", err)
			}

			// Update packages
			return controller.Update(packagesPath, *sshConfig)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "ssh-config.json", "SSH configuration file path")
	return cmd
}
