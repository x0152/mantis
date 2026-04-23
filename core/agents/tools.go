package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/google/uuid"
	robcron "github.com/robfig/cron/v3"

	"mantis/core/protocols"
	"mantis/core/types"
	"mantis/shared"
)

func truncateForLabel(s string, maxRunes int) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" || maxRunes <= 0 {
		return ""
	}
	r := []rune(trimmed)
	if len(r) <= maxRunes {
		return trimmed
	}
	return string(r[:maxRunes]) + "..."
}

func (a *MantisAgent) sshTool(c types.Connection) types.Tool {
	connName := c.Name
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
			task := truncateForLabel(input.Task, 140)
			if task != "" {
				return label + task
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
			selection := a.resolveConnectionModelSelection(connCopy)
			model, err := shared.ResolveModel(ctx, a.modelStore, selection.ModelID)
			if err != nil {
				return "", fmt.Errorf("agent %s: %w", connName, err)
			}
			shared.SetModelMeta(ctx, model.ID, model.Name, selection.PresetID, selection.PresetName, selection.ModelRole)
			var sshCfg SSHConfig
			_ = json.Unmarshal(rawConfig, &sshCfg)

			limits := a.sshAgent.Limits()
			sshCtx := ctx
			if limits.ServerTimeout > 0 {
				var cancel context.CancelFunc
				sshCtx, cancel = context.WithTimeout(ctx, limits.ServerTimeout)
				defer cancel()
			}

			ch, err := a.sshAgent.Execute(sshCtx, SSHInput{
				Model:      model,
				SSHConfig:  sshCfg,
				Connection: connCopy,
				Task:       input.Task,
			})
			if err != nil {
				return "", fmt.Errorf("agent %s: %w", connName, err)
			}
			text, runErr := shared.CollectText(ch)
			if runErr != nil {
				return annotateServerLimit(text, runErr, sshCtx, limits), nil
			}
			return text, nil
		},
	}
}

func annotateServerLimit(partial string, runErr error, ctx context.Context, limits shared.Limits) string {
	marker := ""
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		marker = shared.StopReasonServerTimeout(limits.ServerTimeout)
	case errors.Is(ctx.Err(), context.Canceled):
		marker = shared.StopReasonUser()
	case runErr != nil && strings.Contains(runErr.Error(), "max iterations reached"):
		marker = shared.StopReasonServerIterations(limits.ServerMaxIterations)
	default:
		if partial != "" {
			return partial + "\nerror: " + runErr.Error()
		}
		return "error: " + runErr.Error()
	}
	if partial == "" {
		return marker
	}
	return partial + "\n\n" + marker
}

func (a *MantisAgent) skillTool(c types.Connection, s types.Skill) types.Tool {
	connName := c.Name
	rawConfig := c.Config
	toolName := skillToolName(s)
	params := skillParametersSchema(s.Parameters)
	description := strings.TrimSpace(s.Description)
	toolDescription := fmt.Sprintf("Run skill %q on %s.", s.Name, connName)
	if description != "" {
		toolDescription += " " + description
	}

	return types.Tool{
		Name:        toolName,
		Description: toolDescription,
		Icon:        "skill",
		Label: func(args string) string {
			payload := strings.TrimSpace(args)
			if payload == "" || payload == "{}" {
				return "Skill: " + s.Name
			}
			payload = truncateForLabel(payload, 140)
			return "Skill: " + s.Name + " " + payload
		},
		Parameters: params,
		Execute: func(_ context.Context, args string) (string, error) {
			input := map[string]any{}
			payload := strings.TrimSpace(args)
			if payload != "" && payload != "null" {
				if err := json.Unmarshal([]byte(payload), &input); err != nil {
					return "", err
				}
			}
			script, err := renderSkillScript(s.Script, input)
			if err != nil {
				return "", err
			}
			var sshCfg SSHConfig
			_ = json.Unmarshal(rawConfig, &sshCfg)
			return executeSkillScript(sshCfg, script)
		},
	}
}

func skillParametersSchema(raw json.RawMessage) map[string]any {
	fallback := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
	if len(raw) == 0 {
		return fallback
	}
	payload := strings.TrimSpace(string(raw))
	if payload == "" || payload == "null" {
		return fallback
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil || schema == nil {
		return fallback
	}
	if _, ok := schema["type"]; !ok {
		schema["type"] = "object"
	}
	if _, ok := schema["properties"]; !ok {
		schema["properties"] = map[string]any{}
	}
	return schema
}

func renderSkillScript(source string, args map[string]any) (string, error) {
	tmpl, err := template.New("skill").Option("missingkey=zero").Parse(source)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	if err := tmpl.Execute(&out, args); err != nil {
		return "", err
	}
	script := out.String()
	if strings.TrimSpace(script) == "" {
		return "", fmt.Errorf("skill script is empty")
	}
	return script, nil
}

func executeSkillScript(cfg SSHConfig, script string) (string, error) {
	delimiter := "MANTIS_SKILL_" + strings.ReplaceAll(uuid.New().String(), "-", "")
	for strings.Contains(script, delimiter) {
		delimiter = "MANTIS_SKILL_" + strings.ReplaceAll(uuid.New().String(), "-", "")
	}
	command := fmt.Sprintf("tmp=$(mktemp /tmp/mantis-skill-XXXXXX.sh)\ncat <<'%s' > \"$tmp\"\n%s\n%s\nchmod +x \"$tmp\"\nbash \"$tmp\"\nstatus=$?\nrm -f \"$tmp\"\nexit $status", delimiter, script, delimiter)
	return execSSH(cfg, command)
}

func skillToolName(s types.Skill) string {
	name := normalizeToolPart(s.Name)
	if len(name) > 48 {
		name = strings.Trim(name[:48], "_")
	}
	suffix := normalizeToolPart(s.ID)
	if len(suffix) > 8 {
		suffix = suffix[:8]
	}
	if suffix == "" {
		suffix = strings.ReplaceAll(uuid.New().String(), "-", "")[:8]
	}
	out := "skill_" + name + "_" + suffix
	if len(out) > 64 {
		out = out[:64]
	}
	return strings.Trim(out, "_")
}

func normalizeToolPart(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastUnderscore := false
	for _, r := range value {
		isAlpha := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		if isAlpha || isDigit {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "skill"
	}
	return out
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

// --- Notification tools ---

func (a *MantisAgent) sendNotificationTool() types.Tool {
	return types.Tool{
		Name:        "send_notification",
		Description: "Send a notification message to the user via Telegram. Use for alerts, reports, or delivering important information.",
		Icon:        "bell",
		Label: func(args string) string {
			var input struct {
				Text string `json:"text"`
			}
			_ = json.Unmarshal([]byte(args), &input)
			text := truncateForLabel(input.Text, 140)
			if text != "" {
				return "Notify: " + text
			}
			return "Send notification"
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"text": map[string]any{
					"type":        "string",
					"description": "Notification message text",
				},
			},
			"required": []string{"text"},
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			if a.channelStore == nil {
				return "", fmt.Errorf("channels are not configured")
			}
			var input struct {
				Text string `json:"text"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			text := strings.TrimSpace(input.Text)
			if text == "" {
				return "", fmt.Errorf("text is required")
			}

			channels, err := a.channelStore.List(ctx, types.ListQuery{})
			if err != nil {
				return "", fmt.Errorf("failed to load channels: %w", err)
			}

			var token string
			var chatID int64
			for _, ch := range channels {
				if ch.Type != "telegram" || strings.TrimSpace(ch.Token) == "" {
					continue
				}
				if token == "" {
					token = strings.TrimSpace(ch.Token)
				}
				if len(ch.AllowedUserIDs) > 0 {
					chatID = ch.AllowedUserIDs[0]
					token = strings.TrimSpace(ch.Token)
					break
				}
			}
			if token == "" {
				return "", fmt.Errorf("no telegram channel configured")
			}
			if chatID == 0 {
				return "", fmt.Errorf("no telegram recipient found (set allowedUserIds in telegram channel)")
			}

			if err := sendTelegramMessage(token, chatID, text); err != nil {
				return "", err
			}

			out, _ := json.Marshal(map[string]any{"ok": true, "channel": "telegram", "chatId": chatID})
			return string(out), nil
		},
	}
}

func sendTelegramMessage(token string, chatID int64, text string) error {
	if err := sendTelegramRaw(token, chatID, text, "MarkdownV2"); err != nil {
		if err2 := sendTelegramRaw(token, chatID, text, ""); err2 != nil {
			return fmt.Errorf("markdownv2 failed: %w; plain fallback failed: %w", err, err2)
		}
	}
	return nil
}

func sendTelegramRaw(token string, chatID int64, text, parseMode string) error {
	u := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := map[string]any{"chat_id": chatID, "text": text}
	if parseMode != "" {
		payload["parse_mode"] = parseMode
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(u, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// --- Plan tools ---

func (a *MantisAgent) planListTool() types.Tool {
	return types.Tool{
		Name:        "plan_list",
		Description: "List available agentic workflow plans.",
		Icon:        "git-branch",
		Label:       func(_ string) string { return "List plans" },
		Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
		Execute: func(ctx context.Context, _ string) (string, error) {
			if a.planStore == nil {
				return "", fmt.Errorf("plans are not configured")
			}
			items, err := a.planStore.List(ctx, types.ListQuery{})
			if err != nil {
				return "", err
			}
			if items == nil {
				items = []types.Plan{}
			}
			type planSummary struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description,omitempty"`
				Schedule    string `json:"schedule,omitempty"`
				Enabled     bool   `json:"enabled"`
				Nodes       int    `json:"nodes"`
				HasParams   bool   `json:"hasParams,omitempty"`
			}
			summaries := make([]planSummary, len(items))
			for i, p := range items {
				hasParams := len(p.Parameters) > 0 && string(p.Parameters) != "{}" && string(p.Parameters) != "null"
				summaries[i] = planSummary{
					ID: p.ID, Name: p.Name, Description: p.Description,
					Schedule: p.Schedule, Enabled: p.Enabled, Nodes: len(p.Graph.Nodes),
					HasParams: hasParams,
				}
			}
			out, _ := json.Marshal(map[string]any{"plans": summaries})
			return string(out), nil
		},
	}
}

func (a *MantisAgent) planGetTool() types.Tool {
	return types.Tool{
		Name:        "plan_get",
		Description: "Get full details of a plan by ID, including all steps (nodes), edges, and parameters. Use this to inspect what a plan does before running it.",
		Icon:        "git-branch",
		Label: func(args string) string {
			var input struct {
				ID string `json:"id"`
			}
			_ = json.Unmarshal([]byte(args), &input)
			if input.ID != "" {
				return "Inspect plan " + input.ID[:min(8, len(input.ID))]
			}
			return "Inspect plan"
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "description": "Plan ID"},
			},
			"required": []string{"id"},
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			if a.planStore == nil {
				return "", fmt.Errorf("plans are not configured")
			}
			var input struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			plans, err := a.planStore.Get(ctx, []string{input.ID})
			if err != nil {
				return "", err
			}
			p, ok := plans[input.ID]
			if !ok {
				return "", fmt.Errorf("plan not found: %s", input.ID)
			}
			type nodeDetail struct {
				ID           string `json:"id"`
				Type         string `json:"type"`
				Label        string `json:"label"`
				Prompt       string `json:"prompt"`
				ClearContext bool   `json:"clearContext,omitempty"`
			}
			type edgeDetail struct {
				Source string `json:"source"`
				Target string `json:"target"`
				Label  string `json:"label,omitempty"`
			}
			nodes := make([]nodeDetail, len(p.Graph.Nodes))
			for i, n := range p.Graph.Nodes {
				nodes[i] = nodeDetail{ID: n.ID, Type: string(n.Type), Label: n.Label, Prompt: n.Prompt, ClearContext: n.ClearContext}
			}
			edges := make([]edgeDetail, len(p.Graph.Edges))
			for i, e := range p.Graph.Edges {
				edges[i] = edgeDetail{Source: e.Source, Target: e.Target, Label: e.Label}
			}
			out, _ := json.Marshal(map[string]any{
				"id":          p.ID,
				"name":        p.Name,
				"description": p.Description,
				"schedule":    p.Schedule,
				"enabled":     p.Enabled,
				"parameters":  p.Parameters,
				"nodes":       nodes,
				"edges":       edges,
			})
			return string(out), nil
		},
	}
}

func (a *MantisAgent) planRunTool() types.Tool {
	return types.Tool{
		Name:        "plan_run",
		Description: "Trigger execution of an agentic workflow plan. The plan runs asynchronously and results can be viewed in the Plans UI. If the plan has parameters, pass them in the 'input' object.",
		Icon:        "play",
		Label: func(args string) string {
			var input struct {
				ID string `json:"id"`
			}
			_ = json.Unmarshal([]byte(args), &input)
			if input.ID != "" {
				return "Run plan " + input.ID[:min(8, len(input.ID))]
			}
			return "Run plan"
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{
					"type":        "string",
					"description": "Plan ID to execute",
				},
				"input": map[string]any{
					"type":        "object",
					"description": "Input parameters for the plan (if it requires any)",
				},
			},
			"required": []string{"id"},
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			if a.planRunner == nil {
				return "", fmt.Errorf("plan runner is not configured")
			}
			var input struct {
				ID    string         `json:"id"`
				Input map[string]any `json:"input"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			if strings.TrimSpace(input.ID) == "" {
				return "", fmt.Errorf("plan id is required")
			}
			run, err := a.planRunner.TriggerRun(ctx, input.ID, "chat", input.Input)
			if err != nil {
				return "", err
			}
			out, _ := json.Marshal(map[string]any{
				"ok":     true,
				"run_id": run.ID,
				"status": run.Status,
				"steps":  len(run.Steps),
			})
			return string(out), nil
		},
	}
}

type planStep struct {
	ID     string `json:"id,omitempty"`
	Type   string `json:"type"`
	Label  string `json:"label,omitempty"`
	Prompt string `json:"prompt,omitempty"`
	Yes    string `json:"yes,omitempty"`
	No     string `json:"no,omitempty"`
}

const maxAgentPlanSteps = 15

func stepsToGraph(steps []planStep) (types.PlanGraph, error) {
	if len(steps) == 0 {
		return types.PlanGraph{}, fmt.Errorf("at least one step is required")
	}
	if len(steps) > maxAgentPlanSteps {
		return types.PlanGraph{}, fmt.Errorf("too many steps (max %d)", maxAgentPlanSteps)
	}

	type resolved struct {
		nodeID string
		index  int
	}
	byUserID := map[string]resolved{}

	nodes := make([]types.PlanNode, 0, len(steps))
	for i, s := range steps {
		nid := fmt.Sprintf("n%d", i+1)
		nodeType := types.PlanNodeAction
		if s.Type == "decision" {
			nodeType = types.PlanNodeDecision
		}
		label := s.Label
		if label == "" {
			if nodeType == types.PlanNodeDecision {
				label = fmt.Sprintf("Decision %d", i+1)
			} else {
				label = fmt.Sprintf("Step %d", i+1)
			}
		}

		yPos := 150 * i
		xPos := 250
		pos, _ := json.Marshal(map[string]int{"x": xPos, "y": yPos})

		nodes = append(nodes, types.PlanNode{
			ID:       nid,
			Type:     nodeType,
			Label:    label,
			Prompt:   strings.TrimSpace(s.Prompt),
			Position: pos,
		})

		if s.ID != "" {
			byUserID[s.ID] = resolved{nodeID: nid, index: i}
		}
		byUserID[nid] = resolved{nodeID: nid, index: i}
	}

	resolveTarget := func(ref string, fromIdx int) string {
		if ref == "" || ref == "next" {
			next := fromIdx + 1
			if next < len(nodes) {
				return nodes[next].ID
			}
			return ""
		}
		if ref == "end" || ref == "stop" {
			return ""
		}
		if r, ok := byUserID[ref]; ok {
			return r.nodeID
		}
		return ""
	}

	var edges []types.PlanEdge
	edgeID := 0
	for i, s := range steps {
		node := nodes[i]
		switch node.Type {
		case types.PlanNodeAction:
			target := resolveTarget("next", i)
			if target != "" {
				edgeID++
				edges = append(edges, types.PlanEdge{
					ID: fmt.Sprintf("e%d", edgeID), Source: node.ID, Target: target,
				})
			}
		case types.PlanNodeDecision:
			if yesTarget := resolveTarget(s.Yes, i); yesTarget != "" {
				edgeID++
				edges = append(edges, types.PlanEdge{
					ID: fmt.Sprintf("e%d", edgeID), Source: node.ID, Target: yesTarget, Label: "yes",
				})
			}
			if noTarget := resolveTarget(s.No, i); noTarget != "" {
				edgeID++
				edges = append(edges, types.PlanEdge{
					ID: fmt.Sprintf("e%d", edgeID), Source: node.ID, Target: noTarget, Label: "no",
				})
			}
		}
	}

	return types.PlanGraph{Nodes: nodes, Edges: edges}, nil
}

func (a *MantisAgent) planCreateTool() types.Tool {
	return types.Tool{
		Name:        "plan_create",
		Description: "Create an agentic workflow plan from a list of steps. Each step is either an action (the agent executes a prompt) or a decision (the agent answers yes/no and branches). Plans can optionally have a cron schedule. To modify a plan, delete it and create a new one.",
		Icon:        "git-branch",
		Label: func(args string) string {
			var input struct {
				Name string `json:"name"`
			}
			_ = json.Unmarshal([]byte(args), &input)
			if input.Name != "" {
				return "Create plan: " + input.Name
			}
			return "Create plan"
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name":        map[string]any{"type": "string", "description": "Plan name"},
				"description": map[string]any{"type": "string", "description": "What this plan does (1-2 sentences)"},
				"schedule":    map[string]any{"type": "string", "description": "Cron expression for recurring execution (empty = manual only). Examples: '0 9 * * *' daily at 9am, '*/30 * * * *' every 30 min"},
				"enabled":     map[string]any{"type": "boolean", "description": "Enable the plan (default: true if schedule is set, false otherwise)"},
				"steps": map[string]any{
					"type":        "array",
					"description": "Ordered list of plan steps (max 15)",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"type":   map[string]any{"type": "string", "enum": []string{"action", "decision"}, "description": "Step type"},
							"prompt": map[string]any{"type": "string", "description": "What the agent should do (action) or decide (decision, must be a yes/no question)"},
							"label":  map[string]any{"type": "string", "description": "Short label for the step (optional)"},
							"id":     map[string]any{"type": "string", "description": "Step ID for referencing from decision branches (optional, needed only if a decision targets this step)"},
							"yes":    map[string]any{"type": "string", "description": "Decision only: where to go on YES — 'next' (default), step id, or 'end'"},
							"no":     map[string]any{"type": "string", "description": "Decision only: where to go on NO — 'next', step id, or 'end' (default)"},
						},
						"required": []string{"type", "prompt"},
					},
				},
			},
			"required": []string{"name", "steps"},
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			if a.planStore == nil {
				return "", fmt.Errorf("plan store is not configured")
			}
			var input struct {
				Name        string     `json:"name"`
				Description string     `json:"description"`
				Schedule    string     `json:"schedule"`
				Enabled     *bool      `json:"enabled"`
				Steps       []planStep `json:"steps"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			name := strings.TrimSpace(input.Name)
			if name == "" {
				return "", fmt.Errorf("name is required")
			}
			if len(input.Steps) == 0 {
				return "", fmt.Errorf("at least one step is required")
			}

			schedule := strings.TrimSpace(input.Schedule)
			if schedule != "" {
				parser := robcron.NewParser(robcron.Minute | robcron.Hour | robcron.Dom | robcron.Month | robcron.Dow | robcron.Descriptor)
				if _, err := parser.Parse(schedule); err != nil {
					return "", fmt.Errorf("invalid cron schedule %q: %w", schedule, err)
				}
			}

			graph, err := stepsToGraph(input.Steps)
			if err != nil {
				return "", err
			}

			enabled := schedule != ""
			if input.Enabled != nil {
				enabled = *input.Enabled
			}

			plan := types.Plan{
				ID:          uuid.New().String(),
				Name:        name,
				Description: strings.TrimSpace(input.Description),
				Schedule:    schedule,
				Enabled:     enabled,
				Graph:       graph,
			}
			created, err := a.planStore.Create(ctx, []types.Plan{plan})
			if err != nil {
				return "", err
			}
			p := created[0]
			out, _ := json.Marshal(map[string]any{
				"ok":       true,
				"id":       p.ID,
				"name":     p.Name,
				"schedule": p.Schedule,
				"enabled":  p.Enabled,
				"nodes":    len(p.Graph.Nodes),
				"edges":    len(p.Graph.Edges),
			})
			return string(out), nil
		},
	}
}

func (a *MantisAgent) planUpdateTool() types.Tool {
	return types.Tool{
		Name:        "plan_update",
		Description: "Update plan settings (schedule, enabled, name, description). Use plan_list first to find the plan ID. To change plan steps, delete and recreate the plan.",
		Icon:        "git-branch",
		Label: func(args string) string {
			var input struct {
				Schedule *string `json:"schedule"`
				Enabled  *bool   `json:"enabled"`
			}
			_ = json.Unmarshal([]byte(args), &input)
			if input.Schedule != nil {
				return "Update plan schedule"
			}
			if input.Enabled != nil && !*input.Enabled {
				return "Disable plan"
			}
			if input.Enabled != nil && *input.Enabled {
				return "Enable plan"
			}
			return "Update plan"
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":          map[string]any{"type": "string", "description": "Plan ID"},
				"enabled":     map[string]any{"type": "boolean", "description": "true to enable, false to disable"},
				"schedule":    map[string]any{"type": "string", "description": "Cron expression (e.g. '0 9 * * *') or empty string to remove schedule"},
				"name":        map[string]any{"type": "string", "description": "New plan name"},
				"description": map[string]any{"type": "string", "description": "New plan description"},
			},
			"required": []string{"id"},
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			if a.planStore == nil {
				return "", fmt.Errorf("plan store is not configured")
			}
			var input struct {
				ID          string  `json:"id"`
				Enabled     *bool   `json:"enabled"`
				Schedule    *string `json:"schedule"`
				Name        *string `json:"name"`
				Description *string `json:"description"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			if strings.TrimSpace(input.ID) == "" {
				return "", fmt.Errorf("plan id is required")
			}
			plans, err := a.planStore.Get(ctx, []string{input.ID})
			if err != nil {
				return "", err
			}
			plan, ok := plans[input.ID]
			if !ok {
				return "", fmt.Errorf("plan %q not found", input.ID)
			}
			if input.Enabled != nil {
				plan.Enabled = *input.Enabled
			}
			if input.Schedule != nil {
				plan.Schedule = *input.Schedule
			}
			if input.Name != nil && strings.TrimSpace(*input.Name) != "" {
				plan.Name = strings.TrimSpace(*input.Name)
			}
			if input.Description != nil {
				plan.Description = *input.Description
			}
			updated, err := a.planStore.Update(ctx, []types.Plan{plan})
			if err != nil {
				return "", err
			}
			out, _ := json.Marshal(map[string]any{
				"ok":       true,
				"id":       updated[0].ID,
				"name":     updated[0].Name,
				"schedule": updated[0].Schedule,
				"enabled":  updated[0].Enabled,
			})
			return string(out), nil
		},
	}
}

func (a *MantisAgent) planDeleteTool() types.Tool {
	return types.Tool{
		Name:        "plan_delete",
		Description: "Delete a plan (scheduled task) by ID. Use plan_list first to find the plan ID.",
		Icon:        "git-branch",
		Label:       func(_ string) string { return "Delete plan" },
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "description": "Plan ID to delete"},
			},
			"required": []string{"id"},
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			if a.planStore == nil {
				return "", fmt.Errorf("plan store is not configured")
			}
			var input struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			if strings.TrimSpace(input.ID) == "" {
				return "", fmt.Errorf("plan id is required")
			}
			if err := a.planStore.Delete(ctx, []string{input.ID}); err != nil {
				return "", err
			}
			out, _ := json.Marshal(map[string]any{"ok": true, "id": input.ID})
			return string(out), nil
		},
	}
}

func (a *MantisAgent) planActiveTool() types.Tool {
	return types.Tool{
		Name:        "plan_active",
		Description: "List currently running plan executions.",
		Icon:        "play",
		Label:       func(_ string) string { return "Active runs" },
		Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
		Execute: func(ctx context.Context, _ string) (string, error) {
			if a.planRunner == nil {
				return "", fmt.Errorf("plan runner is not configured")
			}
			runs, err := a.planRunner.ActiveRuns(ctx)
			if err != nil {
				return "", err
			}
			type runSummary struct {
				RunID     string `json:"runId"`
				PlanID    string `json:"planId"`
				Trigger   string `json:"trigger"`
				StartedAt string `json:"startedAt"`
			}
			summaries := make([]runSummary, len(runs))
			for i, r := range runs {
				summaries[i] = runSummary{
					RunID: r.ID, PlanID: r.PlanID, Trigger: r.Trigger,
					StartedAt: r.StartedAt.Format("2006-01-02 15:04:05 UTC"),
				}
			}
			out, _ := json.Marshal(map[string]any{"running": summaries, "count": len(summaries)})
			return string(out), nil
		},
	}
}

func (a *MantisAgent) planStopTool() types.Tool {
	return types.Tool{
		Name:        "plan_stop",
		Description: "Stop (cancel) a running plan execution by run ID. Use plan_active to find the run ID first.",
		Icon:        "git-branch",
		Label: func(args string) string {
			var input struct {
				RunID string `json:"runId"`
			}
			_ = json.Unmarshal([]byte(args), &input)
			if input.RunID != "" {
				return "Stop run " + input.RunID[:min(8, len(input.RunID))]
			}
			return "Stop plan run"
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"runId": map[string]any{"type": "string", "description": "Run ID to cancel"},
			},
			"required": []string{"runId"},
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			if a.planRunner == nil {
				return "", fmt.Errorf("plan runner is not configured")
			}
			var input struct {
				RunID string `json:"runId"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			if strings.TrimSpace(input.RunID) == "" {
				return "", fmt.Errorf("runId is required")
			}
			run, err := a.planRunner.CancelRun(ctx, input.RunID)
			if err != nil {
				return "", err
			}
			out, _ := json.Marshal(map[string]any{"ok": true, "runId": run.ID, "status": run.Status})
			return string(out), nil
		},
	}
}

// --- Artifact tools ---

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
			var input struct {
				ArtifactID string `json:"artifactId"`
			}
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
				return "<file_content>\n" + header + "\n</file_content>", nil
			}
			return "<file_content>\n" + header + "\n\n" + preview + "\n</file_content>", nil
		},
	}
}

func sendFileChatTool(artifacts *shared.ArtifactStore, requestID string) types.Tool {
	return types.Tool{
		Name:        "send_file",
		Description: "Send an artifact (file/image) to the user.",
		Icon:        "download",
		Label: func(args string) string {
			var input struct {
				FileName string `json:"fileName"`
			}
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

func (a *MantisAgent) sendFileTelegramTool(artifacts *shared.ArtifactStore) types.Tool {
	return types.Tool{
		Name:        "send_file",
		Description: "Send an artifact (file/image) to the user.",
		Icon:        "download",
		Label: func(args string) string {
			var input struct {
				FileName string `json:"fileName"`
			}
			json.Unmarshal([]byte(args), &input)
			if input.FileName != "" {
				return "Send: " + input.FileName
			}
			return "Send file"
		},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"artifactId": map[string]any{"type": "string", "description": "Artifact ID to send"},
				"fileName":   map[string]any{"type": "string", "description": "Optional file name (defaults to artifact name)"},
				"caption":    map[string]any{"type": "string", "description": "Optional caption"},
			},
			"required": []string{"artifactId"},
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			var input struct {
				ArtifactID string `json:"artifactId"`
				FileName   string `json:"fileName"`
				Caption    string `json:"caption"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			art, ok := artifacts.Get(input.ArtifactID)
			if !ok {
				return "", fmt.Errorf("artifact %q not found", input.ArtifactID)
			}
			fileName := input.FileName
			if fileName == "" {
				fileName = art.Name
			}

			channels, err := a.channelStore.List(ctx, types.ListQuery{})
			if err != nil {
				return "", fmt.Errorf("failed to load channels: %w", err)
			}
			var token string
			var chatID int64
			for _, ch := range channels {
				if ch.Type != "telegram" || strings.TrimSpace(ch.Token) == "" {
					continue
				}
				if token == "" {
					token = strings.TrimSpace(ch.Token)
				}
				if len(ch.AllowedUserIDs) > 0 {
					chatID = ch.AllowedUserIDs[0]
					token = strings.TrimSpace(ch.Token)
					break
				}
			}
			if token == "" {
				return "", fmt.Errorf("no telegram channel configured")
			}
			if chatID == 0 {
				return "", fmt.Errorf("no telegram recipient found")
			}

			if err := sendTelegramDocument(ctx, token, chatID, fileName, art.Bytes, input.Caption); err != nil {
				return "", err
			}
			out, _ := json.Marshal(map[string]any{"ok": true, "channel": "telegram", "chatId": chatID, "fileName": fileName})
			return string(out), nil
		},
	}
}

func sendTelegramDocument(ctx context.Context, token string, chatID int64, fileName string, data []byte, caption string) error {
	u := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", token)
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("chat_id", fmt.Sprintf("%d", chatID))
	if caption != "" {
		_ = w.WriteField("caption", caption)
	}
	part, err := w.CreateFormFile("document", fileName)
	if err != nil {
		return err
	}
	if _, err := part.Write(data); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", u, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func artifactTranscribeTool(artifacts *shared.ArtifactStore, asr protocols.ASR) types.Tool {
	return types.Tool{
		Name:        "artifact_transcribe",
		Description: "Transcribe an audio artifact to text (speech-to-text).",
		Icon:        "mic",
		Label: func(args string) string {
			var input struct {
				ArtifactID string `json:"artifactId"`
			}
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
			return "<file_content>\n" + strings.TrimSpace(text) + "\n</file_content>", nil
		},
	}
}

func (a *MantisAgent) loadVisionModelID() string {
	if pid := a.loadDefaultPresetID("chat"); pid != "" {
		if p, err := shared.ResolvePreset(context.Background(), a.presetStore, pid); err == nil {
			if id, _ := presetImageModelID(p); id != "" {
				return id
			}
		}
	}
	if pid := a.loadDefaultPresetID("server"); pid != "" {
		if p, err := shared.ResolvePreset(context.Background(), a.presetStore, pid); err == nil {
			if id, _ := presetImageModelID(p); id != "" {
				return id
			}
		}
	}
	return ""
}

func (a *MantisAgent) artifactReadImageTool(artifacts *shared.ArtifactStore) types.Tool {
	return types.Tool{
		Name:        "artifact_read_image",
		Description: "Read an image artifact: extract text (OCR) and describe content (Vision LLM).",
		Icon:        "eye",
		Label: func(args string) string {
			var input struct {
				ArtifactID string `json:"artifactId"`
			}
			json.Unmarshal([]byte(args), &input)
			if input.ArtifactID != "" {
				return "Read image: " + input.ArtifactID[:8]
			}
			return "Read image"
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
			var input struct {
				ArtifactID string `json:"artifactId"`
			}
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}
			art, ok := artifacts.Get(input.ArtifactID)
			if !ok {
				return "", fmt.Errorf("unknown artifact_id: %s", input.ArtifactID)
			}
			format := art.Format
			if format == "" {
				format = strings.TrimPrefix(art.MIME, "image/")
			}
			if format == "" {
				format = "png"
			}

			visionModelID := a.loadVisionModelID()
			if a.ocr == nil && (visionModelID == "" || a.vision == nil) {
				return "", fmt.Errorf("neither OCR nor Vision LLM is configured")
			}

			var ocrText, visionText string
			var ocrErr, visionErr error

			if a.ocr != nil {
				ocrText, ocrErr = a.ocr.ExtractText(ctx, bytes.NewReader(art.Bytes), format)
				ocrText = strings.TrimSpace(ocrText)
			}

			if visionModelID != "" && a.vision != nil {
				model, err := shared.ResolveModel(ctx, a.modelStore, visionModelID)
				if err == nil {
					conn, err := shared.ResolveConnection(ctx, a.llmConnStore, model.ConnectionID)
					if err == nil {
						prompt := "Describe everything visible in this image: objects, text, layout, colors, people, actions, UI elements, charts, diagrams — anything present. Be thorough but concise, no filler."
						if ocrText != "" {
							prompt = "Describe everything visible in this image: objects, text, layout, colors, people, actions, UI elements, charts, diagrams — anything present. Be thorough but concise, no filler. OCR extracted the following text from the image:\n" + ocrText
						}
						visionText, visionErr = a.vision.Describe(ctx, conn.BaseURL, conn.APIKey, model.Name, art.Bytes, format, prompt)
						visionText = strings.TrimSpace(visionText)
					} else {
						visionErr = err
					}
				} else {
					visionErr = err
				}
			}

			var parts []string

			if ocrText != "" {
				parts = append(parts, "--- OCR Text ---\n"+ocrText)
			} else if a.ocr != nil && ocrErr != nil {
				parts = append(parts, "--- OCR ---\nError: "+ocrErr.Error())
			} else if a.ocr == nil {
				parts = append(parts, "Note: OCR is not configured")
			}

			if visionText != "" {
				parts = append(parts, "--- Image Description ---\n"+visionText)
			} else if visionModelID != "" && a.vision != nil && visionErr != nil {
				parts = append(parts, "--- Vision ---\nError: "+visionErr.Error())
			} else if visionModelID == "" || a.vision == nil {
				parts = append(parts, "Note: Vision LLM is not configured (set Image Model in your preset)")
			}

			return "<file_content>\n" + strings.Join(parts, "\n\n") + "\n</file_content>", nil
		},
	}
}

// --- Utility tools ---

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
