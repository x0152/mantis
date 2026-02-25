package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	robcron "github.com/robfig/cron/v3"

	agent "mantis/core/plugins/agent"
	"mantis/core/plugins/guard"
	"mantis/core/protocols"
	"mantis/core/types"
	"mantis/shared"
)

const (
	telegramMarkdownV2ReservedChars = "_ * [ ] ( ) ~ ` > # + - = | { } . !"
	mdBacktick                      = "`"
	mdFence                         = "```"

	mantisBasePrompt = `You are Mantis, a helpful AI assistant that manages remote servers and tools on behalf of the user. Your job is to understand what the user needs, take action quickly, and report back concisely.

Personality:
- Be maximally concise. No filler, no preamble, no "Sure!", no "Great question!". Get straight to the point.
- Be proactive: if you notice something off (errors, warnings, resource issues) while executing a task, flag it without being asked.
- If a request is ambiguous, make your best guess and act — but mention your assumption in one short line so the user can correct you.
- If something fails, explain what went wrong and suggest a fix or next step. Never just say "an error occurred".
- When reporting results, highlight what matters: the answer, the change made, the key numbers. Skip noise.
- Match the user's tone and language. If they write casually, respond casually. If they write in Russian, respond in Russian.

Execution:
- All server actions go through tool calls. Never write shell commands in text instead of calling a tool.
- When calling ssh_* tools, describe the task in plain language (goal + expected result). The SSH agent picks the commands. You are the manager, not the executor.
- Before a tool call, give a one-line heads-up (what and why). After, report the outcome in 1-3 sentences.
- If the task needs multiple steps, chain them without asking for permission at each step. Report the full result at the end.
- NEVER make up factual data (prices, stats, versions, dates, IPs, etc.). If you are not 100% certain, use a tool to check. When the user asks for real-time or factual information, ALWAYS verify via a tool call — even if you just answered a similar question. Your training data is outdated; the only reliable source is a live check.
- If the user's request can be answered purely from general knowledge (concepts, explanations, how-tos) without factual lookups, answer directly.
- You have long-term memory about the user and their servers. Use this knowledge naturally — as if you simply remember it. Never say "according to my notes", "from your profile", "based on stored data", or anything that reveals the memory mechanism.

Tools:

ssh_<server_name> — run a task on a server via SSH agent.
  Parameter task: plain-language description of what to do and what result you expect.
  FORBIDDEN: shell commands, code, or flags in the task parameter.

ssh_download_<server_name> — download a file from the server into a temporary artifact.
  Parameter remotePath: file path on the server.

ssh_upload_<server_name> — upload a temporary artifact to the server.
  Parameters: artifactId, remotePath.

artifacts_list — list temporary in-memory artifacts.

artifact_read_text — preview a text artifact (avoids pulling large files into context).
  Parameter artifactId.

artifact_send_to_chat — queue an artifact for delivery to the user.
  Parameter artifactId.

artifact_transcribe — speech-to-text on an audio artifact.
  Parameter artifactId.

artifact_ocr — OCR on an image artifact.
  Parameter artifactId.

cron_create — create a scheduled job (cron expression + prompt).
  Parameters: schedule, prompt, name (optional).
  The prompt should describe ONLY the task (e.g. "Get the current BTC/USD price"). Delivery channel and recipient are configured globally in settings — never include them in the prompt.

cron_list — list scheduled jobs.

cron_delete — delete a scheduled job by id.

All artifacts are temporary (~30 min TTL, in-memory).

Formatting:
- If reply channel = telegram, use Telegram MarkdownV2.
- If reply channel = web, use Markdown (CommonMark + GFM).
- Otherwise, plain text.

Telegram MarkdownV2:
- bold: *text*, italic: _text_, underline: __text__, strikethrough: ~text~, spoiler: ||text||
- monospace: ` + mdBacktick + `code` + mdBacktick + `, code block: ` + mdFence + `
code
` + mdFence + `
- Escape ` + telegramMarkdownV2ReservedChars + ` with backslash for literal use.
- Lists: use "•" or "—" instead of "-" or "1.".
- When in doubt, plain text.
- Never insert raw file bytes.`
)

type MantisInput struct {
	SessionID string
	ModelID   string
	Content   string
	Artifacts *shared.ArtifactStore
	RequestID string // per-request id (e.g. assistant message id)

	// Request/response context for the LLM (helps it choose formatting and delivery).
	Source       string // e.g. "telegram", "web", "cron"
	ReplyChannel string // e.g. "telegram", "web"
	ReplyTo      string // e.g. telegram chat_id, web session id, etc.

	// If true, do not include message history in the LLM context.
	DisableHistory bool
}

type MantisAgent struct {
	messageStore    protocols.Store[string, types.ChatMessage]
	modelStore      protocols.Store[string, types.Model]
	llmConnStore    protocols.Store[string, types.LlmConnection]
	connectionStore protocols.Store[string, types.Connection]
	cronJobStore    protocols.Store[string, types.CronJob]
	configStore     protocols.Store[string, types.Config]
	agent           *agent.Agent
	sshAgent        *SSHAgent
	asr             protocols.ASR
	ocr             protocols.OCR
}

func NewMantisAgent(
	messageStore protocols.Store[string, types.ChatMessage],
	modelStore protocols.Store[string, types.Model],
	llmConnStore protocols.Store[string, types.LlmConnection],
	connectionStore protocols.Store[string, types.Connection],
	cronJobStore protocols.Store[string, types.CronJob],
	configStore protocols.Store[string, types.Config],
	llm protocols.LLM,
	g *guard.Guard,
	sessionLogger *shared.SessionLogger,
	asr protocols.ASR,
	ocr protocols.OCR,
) *MantisAgent {
	return &MantisAgent{
		messageStore:    messageStore,
		modelStore:      modelStore,
		llmConnStore:    llmConnStore,
		connectionStore: connectionStore,
		cronJobStore:    cronJobStore,
		configStore:     configStore,
		agent:           agent.New(llm),
		sshAgent:        NewSSHAgent(llmConnStore, llm, g, sessionLogger),
		asr:             asr,
		ocr:             ocr,
	}
}

func (a *MantisAgent) Execute(ctx context.Context, in MantisInput) (<-chan types.StreamEvent, error) {
	model, err := shared.ResolveModel(ctx, a.modelStore, in.ModelID)
	if err != nil {
		return nil, err
	}

	conn, err := shared.ResolveConnection(ctx, a.llmConnStore, model.ConnectionID)
	if err != nil {
		return nil, err
	}

	var history []protocols.LLMMessage
	if !in.DisableHistory {
		var err error
		history, err = shared.BuildHistory(ctx, a.messageStore, in.SessionID)
		if err != nil {
			return nil, err
		}
	}

	connections, err := a.connectionStore.List(ctx, types.ListQuery{})
	if err != nil {
		return nil, err
	}

	artifacts := in.Artifacts
	if artifacts == nil {
		artifacts = shared.NewArtifactStore()
	}

	requestID := strings.TrimSpace(in.RequestID)
	if requestID == "" {
		requestID = uuid.New().String()
	}

	tools := a.buildTools(connections, artifacts, requestID)
	prompt := a.buildSystemPrompt(connections, artifacts, in.Source, in.ReplyChannel, in.ReplyTo)

	messages := []protocols.LLMMessage{{Role: "system", Content: prompt}}
	messages = append(messages, history...)
	if in.DisableHistory {
		messages = append(messages, protocols.LLMMessage{Role: "user", Content: in.Content})
	}

	ch, err := a.agent.Execute(ctx, agent.AgentInput{
		LoopInput: agent.LoopInput{
			ActionInput: agent.ActionInput{
				BaseURL:      conn.BaseURL,
				APIKey:       conn.APIKey,
				Model:        model.Name,
				Messages:     messages,
				Tools:        tools,
				ThinkingMode: model.ThinkingMode,
			},
			MaxIterations: 30,
			MessageID:     in.RequestID,
		},
	})
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (a *MantisAgent) loadUserMemories() []string {
	cfgs, err := a.configStore.Get(context.Background(), []string{"default"})
	if err != nil {
		return nil
	}
	cfg, ok := cfgs["default"]
	if !ok {
		return nil
	}
	var data map[string]any
	if err := json.Unmarshal(cfg.Data, &data); err != nil {
		return nil
	}
	enabled, _ := data["memoryEnabled"].(bool)
	if !enabled {
		return nil
	}
	raw, ok := data["userMemories"].([]any)
	if !ok {
		return nil
	}
	var facts []string
	for _, v := range raw {
		if s, ok := v.(string); ok && s != "" {
			facts = append(facts, s)
		}
	}
	return facts
}

func (a *MantisAgent) buildSystemPrompt(connections []types.Connection, artifacts *shared.ArtifactStore, source, replyChannel, replyTo string) string {
	var sb strings.Builder
	sb.WriteString(mantisBasePrompt)
	sb.WriteString(fmt.Sprintf("\n\nCurrent date/time: %s", time.Now().UTC().Format("Monday, 2006-01-02 15:04:05 UTC")))

	if userMem := a.loadUserMemories(); len(userMem) > 0 {
		sb.WriteString("\n\nYou know the following about the user (use naturally, never mention where this knowledge comes from):")
		for _, f := range userMem {
			sb.WriteString("\n- " + f)
		}
	}

	if source != "" || replyChannel != "" || replyTo != "" {
		sb.WriteString("\n\nRequest context:")
		if source != "" {
			sb.WriteString("\n- source: " + source)
		}
		if replyChannel != "" {
			sb.WriteString("\n- reply channel: " + replyChannel)
		}
		if replyTo != "" {
			sb.WriteString("\n- recipient: " + replyTo)
		}
	}

	if artifacts != nil {
		attached := artifacts.List()
		if len(attached) > 0 {
			sb.WriteString("\n\nAvailable temporary artifacts (TTL ~30 min):")
			maxShow := 20
			if len(attached) < maxShow {
				maxShow = len(attached)
			}
			for _, a := range attached[:maxShow] {
				mime := a.MIME
				if mime == "" {
					mime = "unknown"
				}
				format := a.Format
				if format == "" {
					format = "unknown"
				}
				sb.WriteString(fmt.Sprintf("\n- %s (artifact_id=%s, format=%s, mime=%s, size=%d bytes, sha256=%s)",
					a.Name, a.ID, format, mime, a.SizeBytes, a.SHA256,
				))
			}
			if len(attached) > maxShow {
				sb.WriteString(fmt.Sprintf("\n...and %d more", len(attached)-maxShow))
			}
			sb.WriteString("\nUse artifacts_list to see all artifacts and artifact_read_text to inspect contents.")
		}
	}

	if len(connections) > 0 {
		sb.WriteString("\n\nAvailable agents:\n")
		for _, c := range connections {
			sb.WriteString(fmt.Sprintf("\n- %s (%s): %s", c.Name, c.Type, c.Description))
			if len(c.Memories) > 0 {
				sb.WriteString("\n  Notes:")
				for _, m := range c.Memories {
					sb.WriteString(fmt.Sprintf("\n  - %s", m.Content))
				}
			}
		}
	}

	return sb.String()
}

func (a *MantisAgent) buildTools(connections []types.Connection, artifacts *shared.ArtifactStore, requestID string) []types.Tool {
	var tools []types.Tool
	for _, c := range connections {
		switch c.Type {
		case "ssh":
			tools = append(tools, a.sshTool(c))
			tools = append(tools, a.sshDownloadTool(c, artifacts))
			tools = append(tools, a.sshUploadTool(c, artifacts))
		}
	}
	tools = append(tools,
		artifactsListTool(artifacts, requestID),
		artifactReadTextTool(artifacts),
		artifactSendToChatTool(artifacts, requestID),
		artifactTranscribeTool(artifacts, a.asr),
		artifactOCRTool(artifacts, a.ocr),
		a.cronCreateTool(),
		a.cronListTool(),
		a.cronDeleteTool(),
		sumTool(),
	)
	return tools
}

func (a *MantisAgent) sshTool(c types.Connection) types.Tool {
	connName := c.Name
	modelID := c.ModelID
	rawConfig := c.Config
	connCopy := c

	return types.Tool{
		Name:        fmt.Sprintf("ssh_%s", sanitizeName(connName)),
		Description: fmt.Sprintf("Execute tasks on %s via SSH. %s", connName, c.Description),
		Icon:        "terminal",
		Label: func(args string) string {
			var input struct {
				Task string `json:"task"`
			}
			json.Unmarshal([]byte(args), &input)
			label := connName + ": "
			if len(input.Task) > 40 {
				return label + input.Task[:40] + "..."
			}
			if input.Task != "" {
				return label + input.Task
			}
			return label + "task"
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"task": map[string]any{
					"type":        "string",
					"description": fmt.Sprintf("Task to execute on %s", connName),
				},
			},
			"required": []string{"task"},
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			var input struct {
				Task string `json:"task"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			model, err := shared.ResolveModel(ctx, a.modelStore, modelID)
			if err != nil {
				return "", fmt.Errorf("agent %s: %w", connName, err)
			}
			shared.SetModelName(ctx, model.Name)
			var sshCfg SSHConfig
			_ = json.Unmarshal(rawConfig, &sshCfg)
			ch, err := a.sshAgent.Execute(ctx, SSHInput{
				Model:      model,
				SSHConfig:  sshCfg,
				Connection: connCopy,
				Task:       input.Task,
			})
			if err != nil {
				return "", fmt.Errorf("agent %s: %w", connName, err)
			}
			return shared.CollectText(ch)
		},
	}
}

func (a *MantisAgent) sshDownloadTool(c types.Connection, artifacts *shared.ArtifactStore) types.Tool {
	connName := c.Name
	rawConfig := c.Config

	return types.Tool{
		Name:        fmt.Sprintf("ssh_download_%s", sanitizeName(connName)),
		Description: fmt.Sprintf("Download a remote file from %s via SSH into a temporary artifact (available only during this request).", connName),
		Icon:        "download",
		Label: func(args string) string {
			var input struct {
				RemotePath string `json:"remotePath"`
				Name       string `json:"name"`
			}
			_ = json.Unmarshal([]byte(args), &input)
			if input.RemotePath != "" {
				return connName + ": download " + input.RemotePath
			}
			return connName + ": download file"
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"remotePath": map[string]any{
					"type":        "string",
					"description": "Absolute or relative path to the remote file to download",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Optional artifact display name (defaults to the remote file base name)",
				},
			},
			"required": []string{"remotePath"},
		},
		Execute: func(_ context.Context, args string) (string, error) {
			var input struct {
				RemotePath string `json:"remotePath"`
				Name       string `json:"name"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			var sshCfg SSHConfig
			_ = json.Unmarshal(rawConfig, &sshCfg)

			data, err := downloadSSHFile(sshCfg, input.RemotePath, artifacts.MaxFileBytes)
			if err != nil {
				return "", err
			}

			name := input.Name
			if name == "" {
				name = path.Base(input.RemotePath)
			}
			meta, err := artifacts.Put(name, data, "")
			if err != nil {
				return "", err
			}
			out := map[string]any{
				"artifact_id": meta.ID,
				"name":        meta.Name,
				"format":      meta.Format,
				"size_bytes":  meta.SizeBytes,
				"sha256":      meta.SHA256,
			}
			b, _ := json.Marshal(out)
			return string(b), nil
		},
	}
}

func (a *MantisAgent) sshUploadTool(c types.Connection, artifacts *shared.ArtifactStore) types.Tool {
	connName := c.Name
	rawConfig := c.Config

	return types.Tool{
		Name:        fmt.Sprintf("ssh_upload_%s", sanitizeName(connName)),
		Description: fmt.Sprintf("Upload a temporary artifact to %s via SSH (SFTP).", connName),
		Icon:        "download",
		Label: func(args string) string {
			var input struct {
				ArtifactID  string `json:"artifactId"`
				RemotePath  string `json:"remotePath"`
				Overwrite   bool   `json:"overwrite"`
				Permissions string `json:"mode"`
			}
			_ = json.Unmarshal([]byte(args), &input)
			if input.RemotePath != "" {
				return connName + ": upload " + input.RemotePath
			}
			return connName + ": upload file"
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"artifactId": map[string]any{
					"type":        "string",
					"description": "ID of the artifact to upload (from artifacts_list or ssh_download_*)",
				},
				"remotePath": map[string]any{
					"type":        "string",
					"description": "Destination file path on the remote server",
				},
				"overwrite": map[string]any{
					"type":        "boolean",
					"description": "Whether to overwrite the destination file (default: true)",
				},
				"mode": map[string]any{
					"type":        "string",
					"description": "Optional octal permissions, e.g. 0644",
				},
			},
			"required": []string{"artifactId", "remotePath"},
		},
		Execute: func(_ context.Context, args string) (string, error) {
			var input struct {
				ArtifactID  string `json:"artifactId"`
				RemotePath  string `json:"remotePath"`
				Overwrite   *bool  `json:"overwrite"`
				Permissions string `json:"mode"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}

			aData, ok := artifacts.Get(input.ArtifactID)
			if !ok {
				return "", fmt.Errorf("unknown artifact_id: %s", input.ArtifactID)
			}

			overwrite := true
			if input.Overwrite != nil {
				overwrite = *input.Overwrite
			}

			var perm os.FileMode
			if input.Permissions != "" {
				v, err := strconv.ParseUint(input.Permissions, 8, 32)
				if err != nil {
					return "", fmt.Errorf("invalid mode %q: %w", input.Permissions, err)
				}
				perm = os.FileMode(v)
			}

			var sshCfg SSHConfig
			_ = json.Unmarshal(rawConfig, &sshCfg)

			if err := uploadSSHFile(sshCfg, input.RemotePath, aData.Bytes, perm, overwrite); err != nil {
				return "", err
			}
			out := map[string]any{
				"ok":          true,
				"artifact_id": input.ArtifactID,
				"remote_path": input.RemotePath,
			}
			b, _ := json.Marshal(out)
			return string(b), nil
		},
	}
}

func sanitizeName(name string) string {
	r := strings.NewReplacer(" ", "_", "-", "_", ".", "_")
	return strings.ToLower(r.Replace(name))
}

func (a *MantisAgent) cronCreateTool() types.Tool {
	return types.Tool{
		Name:        "cron_create",
		Description: "Create a cron job (schedule + prompt). Schedule: 5 fields (min hour day month weekday), supports @every.",
		Icon:        "clock",
		Label: func(args string) string {
			var input struct {
				Name     string `json:"name"`
				Schedule string `json:"schedule"`
			}
			_ = json.Unmarshal([]byte(args), &input)
			if input.Name != "" {
				return "cron: " + input.Name
			}
			if input.Schedule != "" {
				return "cron: " + input.Schedule
			}
			return "cron: create"
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Job name (optional)",
				},
				"schedule": map[string]any{
					"type":        "string",
					"description": "Cron schedule (min hour day month weekday) or @every",
				},
				"prompt": map[string]any{
					"type":        "string",
					"description": "Task to execute on schedule (just the task, no recipient/channel info)",
				},
				"enabled": map[string]any{
					"type":        "boolean",
					"description": "Whether the job is enabled (default true)",
				},
			},
			"required": []string{"schedule", "prompt"},
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			if a.cronJobStore == nil {
				return "", fmt.Errorf("cron store is not configured")
			}
			var input struct {
				Name     string `json:"name"`
				Schedule string `json:"schedule"`
				Prompt   string `json:"prompt"`
				Enabled  *bool  `json:"enabled"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			if strings.TrimSpace(input.Schedule) == "" {
				return "", fmt.Errorf("schedule is required")
			}
			if strings.TrimSpace(input.Prompt) == "" {
				return "", fmt.Errorf("prompt is required")
			}

			parser := robcron.NewParser(robcron.Minute | robcron.Hour | robcron.Dom | robcron.Month | robcron.Dow | robcron.Descriptor)
			if _, err := parser.Parse(input.Schedule); err != nil {
				return "", fmt.Errorf("invalid cron schedule %q: %w", input.Schedule, err)
			}

			name := strings.TrimSpace(input.Name)
			if name == "" {
				name = "Cron job"
			}
			enabled := true
			if input.Enabled != nil {
				enabled = *input.Enabled
			}

			j := types.CronJob{
				ID:       uuid.New().String(),
				Name:     name,
				Schedule: input.Schedule,
				Prompt:   input.Prompt,
				Enabled:  enabled,
			}
			created, err := a.cronJobStore.Create(ctx, []types.CronJob{j})
			if err != nil {
				return "", err
			}
			out, _ := json.Marshal(created[0])
			return string(out), nil
		},
	}
}

func (a *MantisAgent) cronListTool() types.Tool {
	return types.Tool{
		Name:        "cron_list",
		Description: "List all cron jobs.",
		Icon:        "clock",
		Label:       func(_ string) string { return "List jobs" },
		Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
		Execute: func(ctx context.Context, _ string) (string, error) {
			if a.cronJobStore == nil {
				return "", fmt.Errorf("cron store is not configured")
			}
			items, err := a.cronJobStore.List(ctx, types.ListQuery{})
			if err != nil {
				return "", err
			}
			if items == nil {
				items = []types.CronJob{}
			}
			out, _ := json.Marshal(map[string]any{"jobs": items})
			return string(out), nil
		},
	}
}

func (a *MantisAgent) cronDeleteTool() types.Tool {
	return types.Tool{
		Name:        "cron_delete",
		Description: "Delete a cron job by id.",
		Icon:        "clock",
		Label:       func(_ string) string { return "Delete job" },
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{
					"type":        "string",
					"description": "ID cron-job",
				},
			},
			"required": []string{"id"},
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			if a.cronJobStore == nil {
				return "", fmt.Errorf("cron store is not configured")
			}
			var input struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			if strings.TrimSpace(input.ID) == "" {
				return "", fmt.Errorf("id is required")
			}
			if err := a.cronJobStore.Delete(ctx, []string{input.ID}); err != nil {
				return "", err
			}
			out, _ := json.Marshal(map[string]any{"ok": true, "id": input.ID})
			return string(out), nil
		},
	}
}

func artifactsListTool(artifacts *shared.ArtifactStore, requestID string) types.Tool {
	return types.Tool{
		Name:        "artifacts_list",
		Description: "List temporary file artifacts available during this request.",
		Icon:        "download",
		Label:       func(_ string) string { return "List artifacts" },
		Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
		Execute: func(_ context.Context, _ string) (string, error) {
			out := map[string]any{
				"artifacts": artifacts.List(),
				"outgoing":  artifacts.Outgoing(requestID),
			}
			b, _ := json.Marshal(out)
			return string(b), nil
		},
	}
}

func artifactReadTextTool(artifacts *shared.ArtifactStore) types.Tool {
	return types.Tool{
		Name:        "artifact_read_text",
		Description: "Read a small preview of a temporary artifact as text (for inspection only).",
		Icon:        "eye",
		Label: func(args string) string {
			var input struct{ ArtifactID string `json:"artifactId"` }
			json.Unmarshal([]byte(args), &input)
			if input.ArtifactID != "" {
				return "Read: " + input.ArtifactID[:8]
			}
			return "Read artifact"
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"artifactId": map[string]any{
					"type":        "string",
					"description": "Artifact ID (from artifacts_list)",
				},
				"maxBytes": map[string]any{
					"type":        "integer",
					"description": "Maximum bytes to preview (default: 8192)",
				},
			},
			"required": []string{"artifactId"},
		},
		Execute: func(_ context.Context, args string) (string, error) {
			var input struct {
				ArtifactID string `json:"artifactId"`
				MaxBytes   int    `json:"maxBytes"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			a, ok := artifacts.Get(input.ArtifactID)
			if !ok {
				return "", fmt.Errorf("unknown artifact_id: %s", input.ArtifactID)
			}
			preview := shared.ArtifactInlinePreview(a, input.MaxBytes)
			format := a.Format
			if format == "" {
				format = "unknown"
			}
			mime := a.MIME
			if mime == "" {
				mime = "unknown"
			}
			header := fmt.Sprintf("File: %s (format=%s, mime=%s, %d bytes, sha256=%s)", a.Name, format, mime, a.SizeBytes, a.SHA256)
			if preview == "" {
				return header, nil
			}
			return header + "\n\n" + preview, nil
		},
	}
}

func artifactSendToChatTool(artifacts *shared.ArtifactStore, requestID string) types.Tool {
	return types.Tool{
		Name:        "artifact_send_to_chat",
		Description: "Mark an artifact for delivery to the requester (if the channel supports sending files).",
		Icon:        "download",
		Label: func(args string) string {
			var input struct{ FileName string `json:"fileName"` }
			json.Unmarshal([]byte(args), &input)
			if input.FileName != "" {
				return "Send: " + input.FileName
			}
			return "Send artifact"
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"artifactId": map[string]any{
					"type":        "string",
					"description": "Artifact ID to send",
				},
				"fileName": map[string]any{
					"type":        "string",
					"description": "Optional file name for delivery (defaults to artifact name)",
				},
				"caption": map[string]any{
					"type":        "string",
					"description": "Optional caption",
				},
			},
			"required": []string{"artifactId"},
		},
		Execute: func(_ context.Context, args string) (string, error) {
			var input struct {
				ArtifactID string `json:"artifactId"`
				FileName   string `json:"fileName"`
				Caption    string `json:"caption"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			fileName := input.FileName
			if fileName == "" {
				if a, ok := artifacts.Get(input.ArtifactID); ok {
					fileName = a.Name
				}
			}
			if err := artifacts.MarkToSend(requestID, input.ArtifactID, fileName, input.Caption); err != nil {
				return "", err
			}
			out := map[string]any{
				"ok":          true,
				"request_id":  requestID,
				"artifact_id": input.ArtifactID,
				"file_name":   fileName,
				"note":        "queued for delivery (channel-dependent)",
			}
			b, _ := json.Marshal(out)
			return string(b), nil
		},
	}
}

func artifactTranscribeTool(artifacts *shared.ArtifactStore, asr protocols.ASR) types.Tool {
	return types.Tool{
		Name:        "artifact_transcribe",
		Description: "Transcribe an audio artifact to text (speech-to-text).",
		Icon:        "mic",
		Label: func(args string) string {
			var input struct{ ArtifactID string `json:"artifactId"` }
			json.Unmarshal([]byte(args), &input)
			if input.ArtifactID != "" {
				return "Transcribe: " + input.ArtifactID[:8]
			}
			return "Transcribe audio"
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"artifactId": map[string]any{
					"type":        "string",
					"description": "Artifact ID of the audio file (from artifacts_list)",
				},
			},
			"required": []string{"artifactId"},
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			if asr == nil {
				return "", fmt.Errorf("ASR is not configured")
			}
			var input struct {
				ArtifactID string `json:"artifactId"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			a, ok := artifacts.Get(input.ArtifactID)
			if !ok {
				return "", fmt.Errorf("unknown artifact_id: %s", input.ArtifactID)
			}
			format := a.Format
			if format == "" {
				format = strings.TrimPrefix(a.MIME, "audio/")
			}
			if format == "" {
				format = "ogg"
			}
			text, err := asr.Transcribe(ctx, bytes.NewReader(a.Bytes), format)
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(text), nil
		},
	}
}

func artifactOCRTool(artifacts *shared.ArtifactStore, ocr protocols.OCR) types.Tool {
	return types.Tool{
		Name:        "artifact_ocr",
		Description: "Extract text from an image artifact (OCR).",
		Icon:        "eye",
		Label: func(args string) string {
			var input struct{ ArtifactID string `json:"artifactId"` }
			json.Unmarshal([]byte(args), &input)
			if input.ArtifactID != "" {
				return "OCR: " + input.ArtifactID[:8]
			}
			return "OCR image"
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"artifactId": map[string]any{
					"type":        "string",
					"description": "Artifact ID of the image file (from artifacts_list)",
				},
			},
			"required": []string{"artifactId"},
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			if ocr == nil {
				return "", fmt.Errorf("OCR is not configured")
			}
			var input struct {
				ArtifactID string `json:"artifactId"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			a, ok := artifacts.Get(input.ArtifactID)
			if !ok {
				return "", fmt.Errorf("unknown artifact_id: %s", input.ArtifactID)
			}
			format := a.Format
			if format == "" {
				format = strings.TrimPrefix(a.MIME, "image/")
			}
			if format == "" {
				format = "png"
			}
			text, err := ocr.ExtractText(ctx, bytes.NewReader(a.Bytes), format)
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(text), nil
		},
	}
}

func sumTool() types.Tool {
	return types.Tool{
		Name:        "sum",
		Description: "Calculate the sum of a list of numbers",
		Icon:        "calculator",
		Label: func(args string) string {
			var input struct {
				Numbers []float64 `json:"numbers"`
			}
			json.Unmarshal([]byte(args), &input)
			if len(input.Numbers) > 0 {
				parts := make([]string, len(input.Numbers))
				for i, n := range input.Numbers {
					parts[i] = strconv.FormatFloat(n, 'f', -1, 64)
				}
				return "Sum: " + strings.Join(parts, " + ")
			}
			return "Calculate sum"
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"numbers": map[string]any{
					"type":  "array",
					"items": map[string]any{"type": "number"},
				},
			},
			"required": []string{"numbers"},
		},
		Execute: func(_ context.Context, args string) (string, error) {
			var input struct {
				Numbers []float64 `json:"numbers"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			var total float64
			for _, n := range input.Numbers {
				total += n
			}
			return strconv.FormatFloat(total, 'f', -1, 64), nil
		},
	}
}
