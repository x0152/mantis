package inferenced

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	keyringBackend  = "test"
	defaultBinPath  = "/usr/local/bin/inferenced"
	defaultTimeout  = 30 * time.Second
	walletAccountID = "mantis"
)

var hexLine = regexp.MustCompile(`(?m)^[0-9a-fA-F]{64}$`)

var (
	ErrNotInstalled = errors.New("inferenced binary is not installed on this server")
)

type Wallet struct {
	Address       string   `json:"address"`
	PrivateKeyHex string   `json:"privateKeyHex"`
	Mnemonic      string   `json:"mnemonic"`
	Words         []string `json:"words"`
}

type Runner struct {
	binPath string
	timeout time.Duration
}

func NewRunner(binPath string) *Runner {
	if strings.TrimSpace(binPath) == "" {
		binPath = defaultBinPath
	}
	return &Runner{binPath: binPath, timeout: defaultTimeout}
}

func (r *Runner) BinaryPath() string {
	return r.binPath
}

func (r *Runner) Available() bool {
	if r.binPath == "" {
		return false
	}
	if info, err := os.Stat(r.binPath); err == nil && !info.IsDir() {
		return true
	}
	if _, err := exec.LookPath(r.binPath); err == nil {
		return true
	}
	return false
}

func (r *Runner) Version(ctx context.Context) (string, error) {
	if !r.Available() {
		return "", ErrNotInstalled
	}
	out, err := r.run(ctx, nil, "version")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (r *Runner) CreateWallet(ctx context.Context) (Wallet, error) {
	if !r.Available() {
		return Wallet{}, ErrNotInstalled
	}

	keyringDir, err := os.MkdirTemp("", "mantis-gonka-keyring-*")
	if err != nil {
		return Wallet{}, fmt.Errorf("create keyring dir: %w", err)
	}
	defer os.RemoveAll(keyringDir)

	addOut, err := r.run(ctx, nil,
		"keys", "add", walletAccountID,
		"--keyring-backend", keyringBackend,
		"--keyring-dir", keyringDir,
		"--output", "json",
	)
	if err != nil {
		return Wallet{}, fmt.Errorf("inferenced keys add: %w", err)
	}

	var info struct {
		Address  string `json:"address"`
		Mnemonic string `json:"mnemonic"`
	}
	if err := json.Unmarshal([]byte(addOut), &info); err != nil {
		return Wallet{}, fmt.Errorf("parse inferenced output: %w", err)
	}
	if info.Address == "" || info.Mnemonic == "" {
		return Wallet{}, fmt.Errorf("inferenced returned incomplete wallet data")
	}

	exportOut, err := r.run(ctx, []byte("y\n"),
		"keys", "export", walletAccountID,
		"--keyring-backend", keyringBackend,
		"--keyring-dir", keyringDir,
		"--unarmored-hex", "--unsafe",
	)
	if err != nil {
		return Wallet{}, fmt.Errorf("inferenced keys export: %w", err)
	}
	hex := hexLine.FindString(exportOut)
	if hex == "" {
		hex = strings.TrimSpace(exportOut)
	}
	if len(hex) != 64 {
		return Wallet{}, fmt.Errorf("inferenced returned invalid private key length=%d", len(hex))
	}

	words := strings.Fields(info.Mnemonic)
	return Wallet{
		Address:       info.Address,
		PrivateKeyHex: hex,
		Mnemonic:      info.Mnemonic,
		Words:         words,
	}, nil
}

func (r *Runner) run(ctx context.Context, stdin []byte, args ...string) (string, error) {
	if r.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, r.binPath, args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("%s", errMsg)
	}
	return stdout.String(), nil
}
