package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// UploadFile uploads a local file to the remote server
func (c *Client) UploadFile(localPath, remotePath string) error {
	if c.sftpClient == nil {
		return fmt.Errorf("SFTP client not connected")
	}

	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file %s: %w", localPath, err)
	}
	defer localFile.Close()

	// Create remote file
	remoteFile, err := c.sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	// Copy file content
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// DownloadFile downloads a file from the remote server to local path
func (c *Client) DownloadFile(remotePath, localPath string) error {
	if c.sftpClient == nil {
		return fmt.Errorf("SFTP client not connected")
	}

	// Open remote file
	remoteFile, err := c.sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	// Ensure local directory exists
	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory %s: %w", localDir, err)
	}

	// Create local file
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file %s: %w", localPath, err)
	}
	defer localFile.Close()

	// Copy file content
	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

// ListFiles lists files in a remote directory
func (c *Client) ListFiles(remotePath string) ([]string, error) {
	if c.sftpClient == nil {
		return nil, fmt.Errorf("SFTP client not connected")
	}

	files, err := c.sftpClient.ReadDir(remotePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list remote directory %s: %w", remotePath, err)
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}

	return fileNames, nil
}

// FileExists checks if a file exists on the remote server
func (c *Client) FileExists(remotePath string) (bool, error) {
	if c.sftpClient == nil {
		return false, fmt.Errorf("SFTP client not connected")
	}

	_, err := c.sftpClient.Stat(remotePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check remote file %s: %w", remotePath, err)
	}

	return true, nil
}

// EnsureRemoteDir ensures the remote directory exists
func (c *Client) EnsureRemoteDir(remotePath string) error {
	if c.sftpClient == nil {
		return fmt.Errorf("SFTP client not connected")
	}

	// Check if directory exists
	if _, err := c.sftpClient.Stat(remotePath); err != nil {
		// Directory doesn't exist, create it
		if err := c.sftpClient.MkdirAll(remotePath); err != nil {
			return fmt.Errorf("failed to create remote directory %s: %w", remotePath, err)
		}
	}

	return nil
}

// GetFileSize returns the size of a remote file
func (c *Client) GetFileSize(remotePath string) (int64, error) {
	if c.sftpClient == nil {
		return 0, fmt.Errorf("SFTP client not connected")
	}

	info, err := c.sftpClient.Stat(remotePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get remote file info %s: %w", remotePath, err)
	}

	return info.Size(), nil
}
