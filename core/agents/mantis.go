package agents

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	agent "mantis/core/plugins/agent"
	"mantis/core/plugins/guard"
	"mantis/core/protocols"
	"mantis/core/types"
	"mantis/shared"
)

//go:embed soul.md
var mantisSoul string

const (
	telegramMarkdownV2ReservedChars = "_ * [ ] ( ) ~ ` > # + - = | { } . !"
	mdBacktick                      = "`"
	mdFence                         = "```"

	mantisBasePrompt = `You are Mantis, a helpful AI assistant that manages remote servers and tools on behalf of the user. Your job is to understand what the user needs, take action quickly, and report back concisely.

Personality:
- Be efficient, not robotic. Keep replies tight, but NEVER drop the voice defined in Soul — short status reports, tool-call announcements, and outcome summaries all still carry your tone and character. "Concise" means no filler, NOT no personality.
- Banned openers: "Sure!", "Great question!", "Certainly!", and any localized sycophantic equivalents in other languages. Never start with sycophancy.
- Be proactive: if you notice something off (errors, warnings, resource issues) while executing a task, flag it without being asked.
- If a request is ambiguous, make your best guess and act — but mention your assumption in one short line so the user can correct you.
- If something fails, explain what went wrong and suggest a fix or next step. Never just say "an error occurred".
- When reporting results, highlight what matters: the answer, the change made, the key numbers. Skip noise, keep the tone.
- Match the user's tone and language. If they write casually, respond casually. If they write in Russian, respond in Russian.

Execution:
- All server actions go through tool calls. Never write shell commands in text instead of calling a tool.
- When calling ssh_* tools, describe the task in plain language (goal + expected result). The SSH agent picks the commands. You are the manager, not the executor.
- Before a tool call, give a one-line heads-up (what and why). After, report the outcome in 1-3 sentences.
- If the task needs multiple steps, chain them without asking for permission at each step. Report the full result at the end.
- If a file was already created or generated on a server in a previous step or conversation (scripts, configs, outputs), reuse it — do not recreate it unless the user explicitly asks for a new version or changes.
- NEVER make up factual data (prices, stats, versions, dates, IPs, etc.). If you are not 100% certain, use a tool to check. When the user asks for real-time or factual information, ALWAYS verify via a tool call — even if you just answered a similar question. Your training data is outdated; the only reliable source is a live check.
- If the user's request can be answered purely from general knowledge (concepts, explanations, how-tos) without factual lookups, answer directly.
- You have long-term memory about the user and their servers. Use this knowledge naturally — as if you simply remember it. Never say "according to my notes", "from your profile", "based on stored data", or anything that reveals the memory mechanism.

Tools:

ssh_<server_name> — run a task on a server via SSH agent.
  Parameter task: plain-language description of what to do and what result you expect.
  FORBIDDEN: shell commands, code, or flags in the task parameter.

ssh_runtimectl — runtime controller. Use this to provision a NEW sandbox when the user's request cannot be served by any existing ssh_* connection you already have (e.g. they need rust, node, a specific DB client, a custom toolchain). Ask it in plain language ("need a sandbox with rust + cargo + curl"); it will build, run and register the container and reply with a line READY sb-<name>. On the very next step that sandbox appears in your tool list as ssh_sb_<name> — use it directly for the real workload. ssh_runtimectl itself must not be used to run the user's workload, only to provision.
  Before you call ssh_runtimectl, briefly confirm with the user: creating a sandbox takes ~30-60 seconds, and one of the existing sandboxes may already cover the request — list the existing sandboxes you have and ask whether to reuse one or build a new one. Only skip the confirmation if the user explicitly asked "create a new sandbox".
  Do NOT ask ssh_runtimectl to "list templates" or "read docs" — it is a builder, just describe what you need and it will produce a sandbox.

ssh_download_<server_name> — download a file from the server into a temporary artifact.
  Parameter remotePath: file path on the server.

ssh_upload_<server_name> — upload a temporary artifact to the server.
  Parameters: artifactId, remotePath.

ssh_connection_create — register a NEW remote SSH host as a connection. Use this when the user wants to attach an external machine (their own server, a VPS, a staging box) as opposed to creating a local sandbox. Authenticate with either a password string or a private_key_artifact_id (the artifact id of a PEM private key the user attached to the chat — find it with artifacts_list first). The tool test-dials the host and only saves the connection if the handshake succeeds. The new ssh_<name> tool becomes available on the next assistant turn.
  Before calling, confirm credentials with the user; never guess a password. If the user pasted a private key in chat as a file, look it up via artifacts_list.

ssh_connection_delete — remove a previously registered remote SSH connection by name. Refuses to delete built-in sandboxes; those are managed from the Runtimes page or via mantisctl.

artifacts_list — list temporary in-memory artifacts.

artifact_read_text — preview a text artifact (avoids pulling large files into context).
  Parameter artifactId.

artifact_transcribe — speech-to-text on an audio artifact.
  Parameter artifactId.

artifact_read_image — Read an image: extract text (OCR) and describe content (Vision LLM).
  Parameter artifactId.

send_notification — send a notification message to the user via Telegram.
  Parameter text: message text (supports Telegram MarkdownV2 formatting).
  Use this to deliver alerts, reports, or any important information directly to the user.

plan_list — list all plans (both scheduled tasks and complex workflows). Returns id, name, schedule, enabled status, parameters.

plan_get — get full details of a plan by id: all steps (nodes with prompts), edges, and parameters. Use to inspect what a plan does before running it.
  Parameter id: plan ID.

plan_run — trigger execution of a plan by id. The plan runs asynchronously.
  Parameters: id (plan ID), input (optional object with parameter values if the plan defines parameters).

plan_create — create a multi-step agentic workflow plan.
  Parameters:
    name (string, required) — plan name.
    description (string) — what this plan does.
    schedule (string) — cron expression for recurring execution (empty = manual only).
    enabled (boolean) — default: true if schedule is set, false otherwise.
    steps (array, required) — ordered list of steps (max 15). Each step:
      { "type": "action", "prompt": "task description" } — the agent executes the prompt.
      { "type": "decision", "prompt": "yes/no question", "yes": "target", "no": "target" } — the agent answers and branches.
    Step targets: "next" (default) = next step in order, "end" = stop the plan, or a step "id".
    Add "id" to a step only if a decision needs to jump to it. Add "label" for a short display name.
  Examples:
    Simple scheduled task (1 step): steps=[{"type":"action","prompt":"Check disk usage and send notification"}], schedule="0 9 * * *"
    Multi-step with branching: steps=[{"type":"action","prompt":"Run health check"}, {"type":"decision","prompt":"Any issues found?","yes":"next","no":"ok"}, {"type":"action","prompt":"Send alert about issues"}, {"type":"action","prompt":"Log all clear","id":"ok"}]

plan_update — update plan settings by id. All fields except id are optional — only provided fields are changed.
  Parameters: id (required), enabled (bool), schedule (string, cron expression or empty to remove schedule), name (string), description (string).
  Examples: change schedule: {"id":"...","schedule":"0 */6 * * *"}, disable: {"id":"...","enabled":false}, remove schedule: {"id":"...","schedule":""}.

plan_delete — delete a plan by id. To modify plan steps, delete the plan and create a new one.

plan_active — list currently running plan executions. Returns run IDs, plan IDs, and start times.

plan_stop — cancel a running plan execution by run ID. Use plan_active first to find the run ID.
  Parameter runId.

IMPORTANT: Scheduled tasks, cron jobs, reminders, and plans are all "plans". For any recurring/scheduled task, use plan_create with a schedule. For complex workflows, use plan_create with multiple steps. Use plan_list / plan_update / plan_delete to manage existing plans. Use plan_update to change schedule, enable/disable, or rename. Only use plan_run when the user explicitly asks to run a plan immediately.
NEVER create a plan unless the user explicitly asks to create one. If the user asks to DO something (e.g. "take a screenshot", "check disk"), do it directly — do NOT wrap it in a plan.

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
	presetStore     protocols.Store[string, types.Preset]
	llmConnStore    protocols.Store[string, types.LlmConnection]
	connectionStore protocols.Store[string, types.Connection]
	skillStore      protocols.Store[string, types.Skill]
	planStore       protocols.Store[string, types.Plan]
	planRunner      protocols.PlanRunner
	channelStore    protocols.Store[string, types.Channel]
	settingsStore   protocols.Store[string, types.Settings]
	sessionStore    protocols.Store[string, types.ChatSession]
	agent           *agent.Agent
	sshAgent        *SSHAgent
	runtime         protocols.Runtime
	asr             protocols.ASR
	ocr             protocols.OCR
	vision          protocols.VisionLLM
	limits          shared.Limits
}

func NewMantisAgent(
	messageStore protocols.Store[string, types.ChatMessage],
	modelStore protocols.Store[string, types.Model],
	presetStore protocols.Store[string, types.Preset],
	llmConnStore protocols.Store[string, types.LlmConnection],
	connectionStore protocols.Store[string, types.Connection],
	skillStore protocols.Store[string, types.Skill],
	planStore protocols.Store[string, types.Plan],
	channelStore protocols.Store[string, types.Channel],
	settingsStore protocols.Store[string, types.Settings],
	sessionStore protocols.Store[string, types.ChatSession],
	llm protocols.LLM,
	g *guard.Guard,
	sessionLogger *shared.SessionLogger,
	asr protocols.ASR,
	ocr protocols.OCR,
	vision protocols.VisionLLM,
	limits shared.Limits,
) *MantisAgent {
	return &MantisAgent{
		messageStore:    messageStore,
		modelStore:      modelStore,
		presetStore:     presetStore,
		llmConnStore:    llmConnStore,
		connectionStore: connectionStore,
		skillStore:      skillStore,
		planStore:       planStore,
		channelStore:    channelStore,
		settingsStore:   settingsStore,
		sessionStore:    sessionStore,
		agent:           agent.New(llm),
		sshAgent:        NewSSHAgent(llmConnStore, llm, g, sessionLogger, limits),
		asr:             asr,
		ocr:             ocr,
		vision:          vision,
		limits:          limits,
	}
}

func (a *MantisAgent) Limits() shared.Limits { return a.limits }

func (a *MantisAgent) SetPlanRunner(r protocols.PlanRunner) {
	a.planRunner = r
}

func (a *MantisAgent) SetRuntime(rt protocols.Runtime) {
	a.runtime = rt
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
		history, err = shared.BuildHistory(ctx, a.messageStore, a.sessionStore, in.SessionID)
		if err != nil {
			return nil, err
		}
	}

	connections, err := a.connectionStore.List(ctx, types.ListQuery{})
	if err != nil {
		return nil, err
	}
	var skills []types.Skill
	if a.skillStore != nil {
		skills, err = a.skillStore.List(ctx, types.ListQuery{})
		if err != nil {
			return nil, err
		}
	}
	if skills == nil {
		skills = []types.Skill{}
	}

	artifacts := in.Artifacts
	if artifacts == nil {
		artifacts = shared.NewArtifactStore()
	}

	requestID := strings.TrimSpace(in.RequestID)
	if requestID == "" {
		requestID = uuid.New().String()
	}

	tools := a.buildTools(connections, skills, artifacts, requestID, in.Source)
	prompt := a.buildSystemPrompt(connections, artifacts, in.Source, in.ReplyChannel, in.ReplyTo)

	toolsProvider := func(pctx context.Context) []types.Tool {
		latestConnections, cerr := a.connectionStore.List(pctx, types.ListQuery{})
		if cerr != nil {
			return tools
		}
		var latestSkills []types.Skill
		if a.skillStore != nil {
			s, serr := a.skillStore.List(pctx, types.ListQuery{})
			if serr == nil {
				latestSkills = s
			}
		}
		if latestSkills == nil {
			latestSkills = []types.Skill{}
		}
		return a.buildTools(latestConnections, latestSkills, artifacts, requestID, in.Source)
	}

	messages := []protocols.LLMMessage{{Role: "system", Content: prompt}}
	messages = append(messages, history...)
	if in.DisableHistory {
		messages = append(messages, protocols.LLMMessage{Role: "user", Content: in.Content})
	}

	if in.Source == "plan" {
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "user" {
				messages[i].Content += "\n\n[Execute ONLY this step. Do NOT call extra tools. STOP when done. If failed, reply [ERROR]: reason.]"
				break
			}
		}
	}

	ch, err := a.agent.Execute(ctx, agent.AgentInput{
		LoopInput: agent.LoopInput{
			ActionInput: agent.ActionInput{
				Provider:     conn.Provider,
				BaseURL:      conn.BaseURL,
				APIKey:       conn.APIKey,
				Model:        model.Name,
				Messages:     messages,
				Tools:        tools,
				ThinkingMode: model.ThinkingMode,
			},
			MaxIterations: a.limits.SupervisorMaxIterations,
			MessageID:     in.RequestID,
			ToolsProvider: toolsProvider,
		},
	})
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (a *MantisAgent) loadUserMemories() []string {
	if a.settingsStore == nil {
		return nil
	}
	settingsMap, err := a.settingsStore.Get(context.Background(), []string{"default"})
	if err != nil {
		return nil
	}
	settings, ok := settingsMap["default"]
	if !ok {
		return nil
	}
	if !settings.MemoryEnabled {
		return nil
	}
	var facts []string
	for _, s := range settings.UserMemories {
		if s != "" {
			facts = append(facts, s)
		}
	}
	return facts
}

func (a *MantisAgent) buildSystemPrompt(connections []types.Connection, artifacts *shared.ArtifactStore, source, replyChannel, replyTo string) string {
	var sb strings.Builder
	prompt := mantisBasePrompt
	prompt = strings.Replace(prompt,
		"artifact_read_text — preview a text artifact",
		"send_file — send an artifact (file/image) to the user in the current channel (attaches it to the assistant reply in the chat UI, or replies with the file in telegram).\n  Parameter artifactId, optional fileName, caption.\n\nsend_file_telegram — push an artifact to the user's Telegram chat regardless of where they are talking to you right now. Use this when the user writes from the web chat and asks to receive something in Telegram.\n  Parameter artifactId, optional fileName, caption.\n\nartifact_read_text — preview a text artifact",
		1)
	if source == "plan" {
		prompt = "[PIPELINE MODE] You are executing ONE step of a multi-step pipeline.\n" +
			"STRICT RULES:\n" +
			"1. Execute ONLY the single task described in the user message. STOP after completing it.\n" +
			"2. NEVER call tools beyond what the instruction explicitly requires.\n" +
			"3. If the instruction says \"take a screenshot\" — take the screenshot and STOP. Do NOT download, send, or do anything else.\n" +
			"4. If the instruction says \"download\" — download and STOP. Do NOT send or process further.\n" +
			"5. If the instruction says \"send\" — send and STOP.\n" +
			"6. The pipeline will call you again for the next step. You must NOT anticipate it.\n" +
			"7. If you cannot complete the step, respond with [ERROR]: reason.\n\n" + prompt
	}
	sb.WriteString(prompt)
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
				uploaded := a.CreatedAt.Format("2006-01-02 15:04:05 UTC")
				sb.WriteString(fmt.Sprintf("\n- %s (artifact_id=%s, format=%s, mime=%s, size=%d bytes, uploaded=%s)",
					a.Name, a.ID, format, mime, a.SizeBytes, uploaded,
				))
			}
			if len(attached) > maxShow {
				sb.WriteString(fmt.Sprintf("\n...and %d more", len(attached)-maxShow))
			}
			sb.WriteString("\nUse artifacts_list to see all artifacts and artifact_read_text to inspect contents.")
		}
	}

	if len(connections) > 0 {
		sb.WriteString("\n\nAvailable agents (ALREADY registered and ready to use via ssh_<name>/sandbox tools — do NOT call ssh_connection_create for any of these; doing so will fail with a duplicate):\n")
		for _, c := range connections {
			sb.WriteString(fmt.Sprintf("\n- %s (%s): %s", c.Name, c.Type, c.Description))
		}
	}

	if source != "plan" {
		if soul := strings.TrimSpace(mantisSoul); soul != "" {
			sb.WriteString("\n\n========================\nSOUL (this is WHO you are — read this last, apply it to every reply)\n========================\n")
			sb.WriteString(soul)
			sb.WriteString("\n========================\nEND OF SOUL\n========================\n\n")
			sb.WriteString("FINAL VOICE CHECK before you send ANY reply — including one-line status reports, tool-call announcements, and short factual answers:\n")
			sb.WriteString("- Does this sound like ME (the Soul above) — or like a generic assistant / build log?\n")
			sb.WriteString("- Is there at least ONE beat of personality (a reaction, a dry quip, a warm acknowledgment, an ironic aside)?\n")
			sb.WriteString("- Did I avoid sycophantic openers (\"Sure!\", \"Great question!\", and their translations)?\n")
			sb.WriteString("If any answer is no — rewrite before sending. Voice is NEVER optional, even when the reply is three words long.\n")
		}
	}

	return sb.String()
}

func (a *MantisAgent) buildTools(connections []types.Connection, skills []types.Skill, artifacts *shared.ArtifactStore, requestID, source string) []types.Tool {
	var tools []types.Tool
	connectionByID := make(map[string]types.Connection, len(connections))
	for _, c := range connections {
		connectionByID[c.ID] = c
		switch c.Type {
		case "ssh":
			tools = append(tools, a.sshTool(c))
			tools = append(tools, a.sshDownloadTool(c, artifacts))
			tools = append(tools, a.sshUploadTool(c, artifacts))
		}
	}
	for _, s := range skills {
		conn, ok := connectionByID[s.ConnectionID]
		if !ok || conn.Type != "ssh" {
			continue
		}
		tools = append(tools, a.skillTool(conn, s))
	}
	tools = append(tools,
		artifactsListTool(artifacts, requestID),
		artifactReadTextTool(artifacts),
		artifactTranscribeTool(artifacts, a.asr),
		a.artifactReadImageTool(artifacts),
		a.sendNotificationTool(),
		a.sshConnectionCreateTool(artifacts),
		a.sshConnectionDeleteTool(),
		sumTool(),
	)
	if source == "plan" {
		tools = append(tools, a.sendFileTelegramTool(artifacts))
	} else {
		tools = append(tools, sendFileChatTool(artifacts, requestID))
		tools = append(tools, a.sendFileTelegramExplicitTool(artifacts))
	}
	if source != "plan" {
		tools = append(tools,
			a.planListTool(),
			a.planGetTool(),
			a.planRunTool(),
			a.planCreateTool(),
			a.planUpdateTool(),
			a.planDeleteTool(),
			a.planActiveTool(),
			a.planStopTool(),
		)
	}
	return tools
}

type modelSelection struct {
	ModelID    string
	PresetID   string
	PresetName string
	ModelRole  string // primary | fallback | legacy
}

func (a *MantisAgent) resolveConnectionModelSelection(c types.Connection) modelSelection {
	if c.PresetID != "" {
		if p, err := shared.ResolvePreset(context.Background(), a.presetStore, c.PresetID); err == nil {
			if id, role := presetChatModelID(p); id != "" {
				return modelSelection{ModelID: id, PresetID: p.ID, PresetName: p.Name, ModelRole: role}
			}
		}
	}
	pid := a.loadDefaultPresetID("server")
	if pid == "" {
		pid = a.loadDefaultPresetID("chat")
	}
	if pid != "" {
		if p, err := shared.ResolvePreset(context.Background(), a.presetStore, pid); err == nil {
			if id, role := presetChatModelID(p); id != "" {
				return modelSelection{ModelID: id, PresetID: p.ID, PresetName: p.Name, ModelRole: role}
			}
		}
	}
	return modelSelection{ModelID: strings.TrimSpace(c.ModelID), ModelRole: "legacy"}
}

func presetChatModelID(p types.Preset) (string, string) {
	if strings.TrimSpace(p.ChatModelID) != "" {
		return strings.TrimSpace(p.ChatModelID), "primary"
	}
	if strings.TrimSpace(p.FallbackModelID) != "" {
		return strings.TrimSpace(p.FallbackModelID), "fallback"
	}
	return "", ""
}

func presetImageModelID(p types.Preset) (string, string) {
	if strings.TrimSpace(p.ImageModelID) != "" {
		return strings.TrimSpace(p.ImageModelID), "primary"
	}
	if strings.TrimSpace(p.FallbackModelID) != "" {
		return strings.TrimSpace(p.FallbackModelID), "fallback"
	}
	return "", ""
}

func (a *MantisAgent) loadDefaultPresetID(slot string) string {
	if a.settingsStore == nil {
		return ""
	}
	settingsMap, err := a.settingsStore.Get(context.Background(), []string{"default"})
	if err != nil {
		return ""
	}
	settings, ok := settingsMap["default"]
	if !ok {
		return ""
	}
	switch strings.ToLower(strings.TrimSpace(slot)) {
	case "chat":
		return strings.TrimSpace(settings.ChatPresetID)
	case "server":
		return strings.TrimSpace(settings.ServerPresetID)
	default:
		return ""
	}
}
