package commands

import (
	"fmt"
	"os"

	"github.com/rasadov/package-manager/config"
	"github.com/rasadov/package-manager/internal/controller"
	"github.com/spf13/cobra"
)

func Create() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "create <packet.json>",
		Short: "Create and upload a package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packetPath := args[0]

			// Check if packet file exists
			if _, err := os.Stat(packetPath); os.IsNotExist(err) {
				return fmt.Errorf("packet file not found: %s", packetPath)
			}

			// Load SSH configuration
			sshConfig, err := config.LoadSSHConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load SSH config: %w", err)
			}

			// Create package
			return controller.Create(packetPath, *sshConfig)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "ssh-config.json", "SSH configuration file path")
	return cmd
}
