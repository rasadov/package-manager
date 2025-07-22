package config

import "time"

type SSHConfig struct {
	Host      string        `json:"host"`
	Port      int           `json:"port"`
	Username  string        `json:"username"`
	KeyPath   string        `json:"key_path"`
	Timeout   time.Duration `json:"timeout"`
	RemoteDir string        `json:"remote_dir"`
}
