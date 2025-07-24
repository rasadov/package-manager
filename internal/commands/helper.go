package commands

import (
	"encoding/json"
	"os"

	"github.com/rasadov/package-manager/config"
)

func loadSSHConfig(configPath string) (*config.SSHConfig, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config with environment variables
		return &config.SSHConfig{
			Host:      getEnvOrDefault("PM_SSH_HOST", "localhost"),
			Port:      22,
			Username:  getEnvOrDefault("PM_SSH_USER", "user"),
			KeyPath:   getEnvOrDefault("PM_SSH_KEY", "~/.ssh/id_rsa"),
			RemoteDir: getEnvOrDefault("PM_SSH_REMOTE_DIR", "/var/packages"),
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var sshConfig config.SSHConfig
	if err := json.Unmarshal(data, &sshConfig); err != nil {
		return nil, err
	}

	return &sshConfig, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
