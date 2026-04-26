package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"mantis/core/types"
	"mantis/shared"
)

func (a *MantisAgent) sshConnectionCreateTool(artifacts *shared.ArtifactStore) types.Tool {
	return types.Tool{
		Name: "ssh_connection_create",
		Description: "Register a new remote SSH host as a Connection. Authenticate with either a password or a private key previously attached to the chat (pass its artifact_id). The agent can then talk to the host through the ssh_<name> tool after the next turn.",
		Icon: "key",
		Label: func(args string) string {
			var input struct {
				Name string `json:"name"`
				Host string `json:"host"`
			}
			_ = json.Unmarshal([]byte(args), &input)
			label := strings.TrimSpace(input.Name)
			if label == "" {
				label = "connection"
			}
			if input.Host != "" {
				return fmt.Sprintf("Add SSH %s → %s", label, input.Host)
			}
			return "Add SSH " + label
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Short, unique connection name (letters, digits, dashes). Becomes the suffix of the ssh_<name> tool.",
				},
				"host": map[string]any{
					"type":        "string",
					"description": "Remote hostname or IP address.",
				},
				"port": map[string]any{
					"type":        "integer",
					"description": "SSH port. Defaults to 22.",
				},
				"username": map[string]any{
					"type":        "string",
					"description": "SSH login user.",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "One-line description of what this host is for.",
				},
				"password": map[string]any{
					"type":        "string",
					"description": "Password authentication. Mutually exclusive with private_key_artifact_id.",
				},
				"private_key_artifact_id": map[string]any{
					"type":        "string",
					"description": "Artifact id of a PEM private key the user attached to the chat. Mutually exclusive with password.",
				},
			},
			"required": []string{"name", "host", "username"},
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			var input struct {
				Name                 string `json:"name"`
				Host                 string `json:"host"`
				Port                 int    `json:"port"`
				Username             string `json:"username"`
				Description          string `json:"description"`
				Password             string `json:"password"`
				PrivateKeyArtifactID string `json:"private_key_artifact_id"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			name := strings.TrimSpace(input.Name)
			host := strings.TrimSpace(input.Host)
			username := strings.TrimSpace(input.Username)
			if name == "" || host == "" || username == "" {
				return "", fmt.Errorf("name, host and username are required")
			}
			if err := validateConnectionName(name); err != nil {
				return "", err
			}
			port := input.Port
			if port == 0 {
				port = 22
			}

			password := strings.TrimSpace(input.Password)
			keyArtifactID := strings.TrimSpace(input.PrivateKeyArtifactID)
			if (password == "") == (keyArtifactID == "") {
				return "", fmt.Errorf("exactly one of password or private_key_artifact_id must be provided")
			}

			privateKey := ""
			if keyArtifactID != "" {
				art, ok := artifacts.Get(keyArtifactID)
				if !ok {
					return "", fmt.Errorf("unknown artifact_id: %s", keyArtifactID)
				}
				privateKey = string(art.Bytes)
				if !strings.Contains(privateKey, "PRIVATE KEY") {
					return "", fmt.Errorf("artifact %s does not look like a PEM private key", keyArtifactID)
				}
			}

			cfg := SSHConfig{
				Host:       host,
				Port:       port,
				Username:   username,
				Password:   password,
				PrivateKey: privateKey,
			}
			client, dialErr := dialSSH(cfg, 7*time.Second)
			if dialErr != nil {
				return "", fmt.Errorf("ssh connect failed: %w", dialErr)
			}
			_ = client.Close()

			rawConfig, err := json.Marshal(map[string]any{
				"host":       host,
				"port":       port,
				"username":   username,
				"password":   password,
				"privateKey": privateKey,
			})
			if err != nil {
				return "", err
			}

			existing, err := a.findConnectionByExactName(ctx, name)
			if err != nil {
				return "", err
			}
			if existing != nil && existing.Dockerfile != "" {
				return "", fmt.Errorf("connection %q is a managed sandbox — refuse to overwrite", name)
			}

			var saved types.Connection
			if existing != nil {
				existing.Type = "ssh"
				existing.Config = rawConfig
				if input.Description != "" {
					existing.Description = input.Description
				}
				updated, uerr := a.connectionStore.Update(ctx, []types.Connection{*existing})
				if uerr != nil {
					return "", uerr
				}
				saved = updated[0]
			} else {
				conn := types.Connection{
					ID:            uuid.New().String(),
					Type:          "ssh",
					Name:          name,
					Description:   input.Description,
					Config:        rawConfig,
					ProfileIDs:    []string{"unrestricted"},
					Memories:      []types.Memory{},
					MemoryEnabled: true,
				}
				created, cerr := a.connectionStore.Create(ctx, []types.Connection{conn})
				if cerr != nil {
					return "", cerr
				}
				saved = created[0]
			}

			out := map[string]any{
				"ok":            true,
				"connection_id": saved.ID,
				"name":          saved.Name,
				"host":          host,
				"port":          port,
				"auth":          authMethodLabel(password, privateKey),
				"tool":          "ssh_" + sanitizeName(saved.Name),
				"note":          "The ssh_<name> tool becomes available on the next assistant turn.",
			}
			b, _ := json.Marshal(out)
			return string(b), nil
		},
	}
}

func (a *MantisAgent) sshConnectionDeleteTool() types.Tool {
	return types.Tool{
		Name:        "ssh_connection_delete",
		Description: "Delete a previously registered remote SSH Connection by name. Refuses to delete built-in sandboxes (connections that have a stored Dockerfile).",
		Icon:        "trash",
		Label: func(args string) string {
			var input struct {
				Name string `json:"name"`
			}
			_ = json.Unmarshal([]byte(args), &input)
			if input.Name == "" {
				return "Remove SSH connection"
			}
			return "Remove SSH " + input.Name
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Connection name to delete.",
				},
			},
			"required": []string{"name"},
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			var input struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			name := strings.TrimSpace(input.Name)
			if name == "" {
				return "", fmt.Errorf("name is required")
			}
			conn, err := a.findConnectionByExactName(ctx, name)
			if err != nil {
				return "", err
			}
			if conn == nil {
				return "", fmt.Errorf("connection %q not found", name)
			}
			if conn.Dockerfile != "" {
				return "", fmt.Errorf("connection %q is a managed sandbox — delete it from the Runtimes page or with mantisctl instead", name)
			}
			if err := a.connectionStore.Delete(ctx, []string{conn.ID}); err != nil {
				return "", err
			}
			out := map[string]any{
				"ok":            true,
				"connection_id": conn.ID,
				"name":          conn.Name,
			}
			b, _ := json.Marshal(out)
			return string(b), nil
		},
	}
}

func validateConnectionName(name string) error {
	if len(name) == 0 || len(name) > 48 {
		return fmt.Errorf("connection name must be 1-48 characters")
	}
	for _, r := range name {
		ok := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
		if !ok {
			return fmt.Errorf("connection name may contain only letters, digits, dashes and underscores")
		}
	}
	return nil
}

func authMethodLabel(password, privateKey string) string {
	switch {
	case privateKey != "":
		return "private_key"
	case password != "":
		return "password"
	default:
		return "none"
	}
}

func (a *MantisAgent) findConnectionByExactName(ctx context.Context, name string) (*types.Connection, error) {
	items, err := a.connectionStore.List(ctx, types.ListQuery{Filter: map[string]string{"name": name}, Page: types.Page{Limit: 1}})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}
