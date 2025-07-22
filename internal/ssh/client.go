package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"github.com/rasadov/package-manager/config"
	"golang.org/x/crypto/ssh"
)

// Client wraps SSH and SFTP clients
type Client struct {
	config     config.SSHConfig
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

// NewClient creates a new SSH client
func NewClient(config config.SSHConfig) *Client {
	if config.Port == 0 {
		config.Port = 22
	}
	if config.RemoteDir == "" {
		config.RemoteDir = "/var/packages"
	}
	return &Client{config: config}
}

// Connect establishes SSH and SFTP connections
func (c *Client) Connect() error {
	// Load private key
	key, err := c.loadPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to load key: %w", err)
	}

	// SSH config
	sshConfig := &ssh.ClientConfig{
		User:            c.config.Username,
		Auth:            []ssh.AuthMethod{key},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Connect
	address := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	sshClient, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	c.sshClient = sshClient

	// SFTP client
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		c.sshClient.Close()
		return fmt.Errorf("SFTP client failed: %w", err)
	}
	c.sftpClient = sftpClient

	return nil
}

// Close closes connections
func (c *Client) Close() error {
	if c.sftpClient != nil {
		c.sftpClient.Close()
	}
	if c.sshClient != nil {
		c.sshClient.Close()
	}
	return nil
}

// GetSFTPClient returns SFTP client for file operations
func (c *Client) GetSFTPClient() *sftp.Client {
	return c.sftpClient
}

// GetRemoteDir returns remote directory path
func (c *Client) GetRemoteDir() string {
	return c.config.RemoteDir
}

// loadPrivateKey loads SSH private key
func (c *Client) loadPrivateKey() (ssh.AuthMethod, error) {
	keyPath := c.config.KeyPath
	if keyPath[0] == '~' {
		homeDir, _ := os.UserHomeDir()
		keyPath = filepath.Join(homeDir, keyPath[1:])
	}

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}

	return ssh.PublicKeys(signer), nil
}
