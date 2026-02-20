package agents

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/sftp"
)

func downloadSSHFile(cfg SSHConfig, remotePath string, maxBytes int64) ([]byte, error) {
	if remotePath == "" {
		return nil, fmt.Errorf("remote_path is required")
	}
	if maxBytes <= 0 {
		maxBytes = 10 * 1024 * 1024
	}

	client, err := dialSSH(cfg, 15*time.Second)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, fmt.Errorf("sftp: %w", err)
	}
	defer sftpClient.Close()

	f, err := sftpClient.Open(remotePath)
	if err != nil {
		return nil, fmt.Errorf("open remote file: %w", err)
	}
	defer f.Close()

	if st, err := f.Stat(); err == nil {
		if st.Size() > maxBytes {
			return nil, fmt.Errorf("remote file too large: %d bytes (max %d)", st.Size(), maxBytes)
		}
	}

	// Double-check the size during read as well.
	data, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read remote file: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("remote file too large: %d bytes (max %d)", len(data), maxBytes)
	}
	return data, nil
}

func uploadSSHFile(cfg SSHConfig, remotePath string, data []byte, perm os.FileMode, overwrite bool) error {
	if remotePath == "" {
		return fmt.Errorf("remote_path is required")
	}

	client, err := dialSSH(cfg, 15*time.Second)
	if err != nil {
		return err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("sftp: %w", err)
	}
	defer sftpClient.Close()

	flags := os.O_WRONLY | os.O_CREATE
	if overwrite {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}

	f, err := sftpClient.OpenFile(remotePath, flags)
	if err != nil {
		return fmt.Errorf("open remote file for write: %w", err)
	}
	defer f.Close()

	n, err := f.Write(data)
	if err != nil {
		return fmt.Errorf("write remote file: %w", err)
	}
	if n != len(data) {
		return fmt.Errorf("short write: %d/%d", n, len(data))
	}

	if perm != 0 {
		if err := sftpClient.Chmod(remotePath, perm); err != nil {
			return fmt.Errorf("chmod remote file: %w", err)
		}
	}

	return nil
}
