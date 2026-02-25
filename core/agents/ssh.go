package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	agent "mantis/core/plugins/agent"
	"mantis/core/plugins/guard"
	"mantis/core/protocols"
	"mantis/core/types"
	"mantis/shared"
)

const sshBasePrompt = `You are an SSH agent. All actions go through execute_command tool calls only.

Rules:
- Be concise: short answers, no filler, keep full info. Verbose only if user asks.
- One command per call. Explain briefly before each call.
- Verify before acting (which, cat, ls).
- Summarize the result at the end.
- Plain text only, no Markdown/HTML.
- If a command is blocked, do not retry it — use an alternative or inform the user.

execute_command(command: string) — run a shell command on the remote server via SSH.`

type SSHConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	PrivateKey string `json:"privateKey"`
}

type SSHInput struct {
	Model      types.Model
	SSHConfig  SSHConfig
	Connection types.Connection
	Task       string
}

type SSHAgent struct {
	llmConnStore  protocols.Store[string, types.LlmConnection]
	agent         *agent.Agent
	guard         *guard.Guard
	sessionLogger *shared.SessionLogger
}

func NewSSHAgent(llmConnStore protocols.Store[string, types.LlmConnection], llm protocols.LLM, g *guard.Guard, sessionLogger *shared.SessionLogger) *SSHAgent {
	return &SSHAgent{
		llmConnStore:  llmConnStore,
		agent:         agent.New(llm),
		guard:         g,
		sessionLogger: sessionLogger,
	}
}

func (a *SSHAgent) Execute(ctx context.Context, in SSHInput) (<-chan types.StreamEvent, error) {
	conn, err := shared.ResolveConnection(ctx, a.llmConnStore, in.Model.ConnectionID)
	if err != nil {
		return nil, err
	}

	hostReadme, err := a.probeHost(in.SSHConfig)
	if err != nil {
		return nil, fmt.Errorf("ssh probe %s:%d: %w", in.SSHConfig.Host, in.SSHConfig.Port, err)
	}

	prompt := a.buildPrompt(ctx, in.Connection, hostReadme)
	tools := sshTools(in.SSHConfig, a.guard, in.Connection.ProfileIDs)

	messages := []protocols.LLMMessage{
		{Role: "system", Content: prompt},
		{Role: "user", Content: in.Task},
	}

	ch, err := a.agent.Execute(ctx, agent.AgentInput{
		LoopInput: agent.LoopInput{
			ActionInput: agent.ActionInput{
				BaseURL:      conn.BaseURL,
				APIKey:       conn.APIKey,
				Model:        in.Model.Name,
				Messages:     messages,
				Tools:        tools,
				ThinkingMode: in.Model.ThinkingMode,
			},
			MaxIterations: 30,
		},
	})
	if err != nil {
		return nil, err
	}

	if a.sessionLogger != nil {
		ch = a.sessionLogger.Wrap(ctx, in.Connection.ID, "ssh", in.Task, ch)
	}

	return ch, nil
}

func (a *SSHAgent) probeHost(cfg SSHConfig) (string, error) {
	client, err := dialSSH(cfg, 10*time.Second)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("ssh session: %w", err)
	}
	defer session.Close()

	var stdout bytes.Buffer
	session.Stdout = &stdout
	_ = session.Run("cat ~/README.md 2>/dev/null || cat /etc/mantis/README.md 2>/dev/null")

	return strings.TrimSpace(stdout.String()), nil
}

func (a *SSHAgent) buildPrompt(ctx context.Context, c types.Connection, hostReadme string) string {
	var sb strings.Builder
	sb.WriteString(sshBasePrompt)
	sb.WriteString(fmt.Sprintf("\n\nCurrent date/time: %s", time.Now().UTC().Format("Monday, 2006-01-02 15:04:05 UTC")))

	if c.Description != "" {
		sb.WriteString(fmt.Sprintf("\n\nServer: %s\nDescription: %s", c.Name, c.Description))
	}

	if hostReadme != "" {
		sb.WriteString("\n\n--- Host instruction (README.md) ---\n")
		sb.WriteString(hostReadme)
		sb.WriteString("\n--- End of instruction ---")
	}

	if len(c.Memories) > 0 {
		sb.WriteString("\n\nYou already know about this server:")
		for _, m := range c.Memories {
			sb.WriteString(fmt.Sprintf("\n- %s", m.Content))
		}
	}

	if desc := a.guard.Describe(ctx, c.ProfileIDs); desc != "" {
		sb.WriteString("\n\n")
		sb.WriteString(desc)
	}

	return sb.String()
}

func sshTools(cfg SSHConfig, g *guard.Guard, profileIDs []string) []types.Tool {
	return []types.Tool{
		{
			Name:        "execute_command",
			Description: "Execute a shell command on the remote server via SSH",
			Icon:        "terminal",
			Label: func(args string) string {
				var input struct {
					Command string `json:"command"`
				}
				json.Unmarshal([]byte(args), &input)
				if input.Command != "" {
					return "$ " + input.Command
				}
				return "SSH command"
			},
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{
						"type":        "string",
						"description": "Shell command to execute",
					},
				},
				"required": []string{"command"},
			},
			Execute: func(ctx context.Context, args string) (string, error) {
				var input struct {
					Command string `json:"command"`
				}
				if err := json.Unmarshal([]byte(args), &input); err != nil {
					return "", err
				}
			if v := g.Execute(ctx, profileIDs, input.Command); v != nil {
				return fmt.Sprintf("[BLOCKED] %s", v.Message), nil
			}
				return execSSH(cfg, input.Command)
			},
		},
	}
}

func dialSSH(cfg SSHConfig, timeout time.Duration) (*ssh.Client, error) {
	authMethods := []ssh.AuthMethod{}
	if cfg.Password != "" {
		authMethods = append(authMethods, ssh.Password(cfg.Password))
	}
	if cfg.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(cfg.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	sshConfig := &ssh.ClientConfig{
		User:            cfg.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	port := cfg.Port
	if port == 0 {
		port = 22
	}
	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", port))

	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("ssh connect %s: %w", addr, err)
	}
	return client, nil
}

const maxOutputBytes = 32768

func execSSH(cfg SSHConfig, command string) (string, error) {
	client, err := dialSSH(cfg, 10*time.Second)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("ssh session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)
	output := stdout.String()
	if stderr.Len() > 0 {
		output += stderr.String()
	}
	if len(output) > maxOutputBytes {
		total := len(output)
		output = output[:maxOutputBytes] + fmt.Sprintf(
			"\n\n[TRUNCATED: %d/%d bytes shown. Redirect to file and use grep/head/tail.]",
			maxOutputBytes, total)
	}
	if err != nil {
		return output + "\nexit: " + err.Error(), nil
	}
	return output, nil
}
