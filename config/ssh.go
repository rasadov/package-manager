package config

import (
	"encoding/json"
	"os"
	"time"
)

type SSHConfig struct {
	Host      string        `json:"host"`
	Port      int           `json:"port"`
	Username  string        `json:"username"`
	KeyPath   string        `json:"key_path"`
	Timeout   time.Duration `json:"timeout"`
	RemoteDir string        `json:"remote_dir"`
}

func LoadSSHConfig(configPath string) (*SSHConfig, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config with environment variables
		return &SSHConfig{
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

	var sshConfig SSHConfig
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
