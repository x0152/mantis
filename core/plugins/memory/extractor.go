package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"mantis/core/plugins/pipeline"
	"mantis/core/protocols"
	"mantis/core/types"
	"mantis/shared"
)

const userPrompt = `You manage a long-term memory system. You store facts the user wants remembered.
You receive existing facts and a recent conversation.

SAVE:
- Identity: name, language, role, company, timezone, location
- Preferences: tools, formats, styles, workflows
- Projects and goals the user is working on
- Anything the user explicitly asks to remember ("запиши", "запомни", "save this", etc.) — save the ACTUAL content, not a description of the request
- Specific knowledge the user shares in their OWN words: warnings, conclusions, decisions

DO NOT SAVE:
- Anything inside <file_content>...</file_content> tags — these are tool-extracted data (OCR, image descriptions, file previews, transcriptions), NOT user knowledge
- Product specs, labels, or metadata extracted from images — the user just shared a file, they didn't state these as personal facts
- Server/infrastructure details (that goes to server memory)
- Anything the assistant said that the user did not explicitly confirm or state themselves

CRITICAL: save the actual information, not meta-descriptions.
BAD: "likes to track prices of things" — this is a meta-description of behavior.
GOOD: "item X costs $50, item Y is unreliable" — this is the actual fact.

Time-sensitive facts (prices, rates, versions, stats) MUST include the date. If the date is unknown, do not save them.
BAD: "item X costs $50"
GOOD: "item X costs $50 (as of 2025-02-15)"

- REMOVE only facts the conversation explicitly contradicts or the user asks to forget.
- Do NOT remove facts just because they aren't mentioned.

Return strictly valid JSON:
{"add": [], "remove": []}
Example: {"add": ["uses Go and React for main project"], "remove": []}`

const connectionPrompt = `You manage a long-term memory system about a remote server.
You receive existing facts and recent SSH command history.

Worth remembering (only if clearly evident from the history):
- Installed or removed software/packages
- Config changes: edited files, changed values, new configs created
- Services: started, stopped, enabled, created
- Important paths: project dirs, config locations, log paths the user works with
- Problems found: broken configs, recurring errors, permission issues, failed upgrades
- Workarounds applied: if something didn't work and a workaround was used, save it so you don't repeat the debugging next time
- State changes after commands: if a command changed the server state in a meaningful way (new cron job, firewall rule, user created, etc.)
- Architecture: what runs on this server, how it connects to other services

NEVER save: disk/memory/cpu stats, process lists, file contents, query results, network info, uptime, log tails, or anything that changes on every check.

It is completely fine to return empty results. Most sessions have nothing worth adding.
Do NOT force facts out of routine checks — only save when something genuinely new or important happened.

- REMOVE only facts the history explicitly shows are no longer true.
- Do NOT remove facts just because they aren't mentioned.

Return strictly valid JSON:
{"add": [], "remove": []}
Only add when truly warranted: {"add": ["certbot renewal failing due to port 80 blocked by nginx, using dns challenge as workaround"], "remove": []}`

type Extractor struct {
	llm             protocols.LLM
	configStore     protocols.Store[string, types.Config]
	connectionStore protocols.Store[string, types.Connection]
	modelStore      protocols.Store[string, types.Model]
	llmConnStore    protocols.Store[string, types.LlmConnection]
}

func NewExtractor(
	llm protocols.LLM,
	configStore protocols.Store[string, types.Config],
	connectionStore protocols.Store[string, types.Connection],
	modelStore protocols.Store[string, types.Model],
	llmConnStore protocols.Store[string, types.LlmConnection],
) *Extractor {
	return &Extractor{
		llm:             llm,
		configStore:     configStore,
		connectionStore: connectionStore,
		modelStore:      modelStore,
		llmConnStore:    llmConnStore,
	}
}

func (e *Extractor) Extract(ctx context.Context, userContent, assistantContent string, sshSteps []pipeline.SSHStep) {
	cfg, modelID, userFacts, ok := e.loadConfig(ctx)
	if !ok || modelID == "" {
		return
	}

	model, err := shared.ResolveModel(ctx, e.modelStore, modelID)
	if err != nil {
		log.Printf("memory: resolve model: %v", err)
		return
	}
	llmConn, err := shared.ResolveConnection(ctx, e.llmConnStore, model.ConnectionID)
	if err != nil {
		log.Printf("memory: resolve llm connection: %v", err)
		return
	}

	e.extractUser(ctx, llmConn, model, cfg, userFacts, userContent, assistantContent)
	e.extractConnections(ctx, llmConn, model, sshSteps)
}

func (e *Extractor) extractUser(ctx context.Context, llmConn types.LlmConnection, model types.Model, cfg types.Config, existing []string, userContent, assistantContent string) {
	if userContent == "" || assistantContent == "" {
		return
	}

	existingJSON, _ := json.Marshal(existing)
	now := time.Now().UTC().Format("Monday, 2006-01-02 15:04:05 UTC")
	input := fmt.Sprintf("Current date/time: %s\n\nExisting facts:\n%s\n\nConversation:\nUser: %s\nAssistant: %s", now, string(existingJSON), userContent, assistantContent)

	result, err := e.callLLM(ctx, llmConn, model, userPrompt, input)
	if err != nil {
		log.Printf("memory: user extract: %v", err)
		return
	}

	var diff memoryDiff
	if err := json.Unmarshal([]byte(result), &diff); err != nil {
		log.Printf("memory: parse user diff: %v (raw: %s)", err, result)
		return
	}

	if len(diff.Add) == 0 && len(diff.Remove) == 0 {
		return
	}
	e.saveUserFacts(ctx, cfg, mergeFacts(existing, diff))
}

func (e *Extractor) extractConnections(ctx context.Context, llmConn types.LlmConnection, model types.Model, sshSteps []pipeline.SSHStep) {
	if len(sshSteps) == 0 {
		return
	}

	connections, err := e.connectionStore.List(ctx, types.ListQuery{})
	if err != nil {
		log.Printf("memory: list connections: %v", err)
		return
	}

	toolToConn := map[string]types.Connection{}
	for _, c := range connections {
		toolToConn["ssh_"+sanitizeName(c.Name)] = c
	}

	byConn := map[string][]pipeline.SSHStep{}
	connMap := map[string]types.Connection{}
	for _, s := range sshSteps {
		c, ok := toolToConn[s.ToolName]
		if !ok || !c.MemoryEnabled {
			continue
		}
		byConn[c.ID] = append(byConn[c.ID], s)
		connMap[c.ID] = c
	}

	for connID, steps := range byConn {
		c := connMap[connID]

		existingFacts := make([]string, len(c.Memories))
		for i, m := range c.Memories {
			existingFacts[i] = m.Content
		}
		existingJSON, _ := json.Marshal(existingFacts)

		var history strings.Builder
		for _, s := range steps {
			history.WriteString(fmt.Sprintf("Task: %s\nOutput: %s\n\n", s.Task, truncate(s.Result, 2000)))
		}

		nowStr := time.Now().UTC().Format("Monday, 2006-01-02 15:04:05 UTC")
		input := fmt.Sprintf("Current date/time: %s\n\nExisting facts:\n%s\n\nSSH history:\n%s", nowStr, string(existingJSON), history.String())

		result, err := e.callLLM(ctx, llmConn, model, connectionPrompt, input)
		if err != nil {
			log.Printf("memory: connection %s extract: %v", connID, err)
			continue
		}

		var diff memoryDiff
		if err := json.Unmarshal([]byte(result), &diff); err != nil {
			log.Printf("memory: parse connection %s diff: %v (raw: %s)", connID, err, result)
			continue
		}

		if len(diff.Add) == 0 && len(diff.Remove) == 0 {
			continue
		}

		removeSet := map[string]bool{}
		for _, r := range diff.Remove {
			removeSet[r] = true
		}

		seen := map[string]bool{}
		var kept []types.Memory
		for _, m := range c.Memories {
			if !removeSet[m.Content] {
				seen[m.Content] = true
				kept = append(kept, m)
			}
		}

		now := time.Now()
		base := now.UnixMilli()
		for i, f := range diff.Add {
			if !seen[f] {
				seen[f] = true
				kept = append(kept, types.Memory{
					ID:        fmt.Sprintf("%s-%d", connID[:8], base+int64(i)),
					Content:   f,
					CreatedAt: now,
				})
			}
		}

		c.Memories = kept
		if _, err := e.connectionStore.Update(ctx, []types.Connection{c}); err != nil {
			log.Printf("memory: save connection %s: %v", connID, err)
		}
	}
}

func (e *Extractor) callLLM(ctx context.Context, llmConn types.LlmConnection, model types.Model, systemPrompt, userInput string) (string, error) {
	messages := []protocols.LLMMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userInput},
	}

	stream, err := e.llm.ChatStream(ctx, llmConn.BaseURL, llmConn.APIKey, messages, model.Name, nil, "skip")
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for event := range stream {
		if event.Type == "text" {
			sb.WriteString(event.Delta)
		}
	}

	raw := strings.TrimSpace(sb.String())
	if idx := strings.Index(raw, "{"); idx >= 0 {
		raw = raw[idx:]
		if end := strings.LastIndex(raw, "}"); end >= 0 {
			raw = raw[:end+1]
		}
	}
	return raw, nil
}

type memoryDiff struct {
	Add    []string `json:"add"`
	Remove []string `json:"remove"`
}

func mergeFacts(existing []string, diff memoryDiff) []string {
	removeSet := map[string]bool{}
	for _, r := range diff.Remove {
		removeSet[r] = true
	}
	seen := map[string]bool{}
	var result []string
	for _, f := range existing {
		if !removeSet[f] {
			seen[f] = true
			result = append(result, f)
		}
	}
	for _, f := range diff.Add {
		if !seen[f] {
			seen[f] = true
			result = append(result, f)
		}
	}
	return result
}

func (e *Extractor) loadConfig(ctx context.Context) (types.Config, string, []string, bool) {
	cfgs, err := e.configStore.Get(ctx, []string{"default"})
	if err != nil {
		log.Printf("memory: load config: %v", err)
		return types.Config{}, "", nil, false
	}
	cfg, ok := cfgs["default"]
	if !ok {
		return types.Config{}, "", nil, false
	}
	var data map[string]any
	if err := json.Unmarshal(cfg.Data, &data); err != nil {
		return cfg, "", nil, false
	}
	enabled, _ := data["memoryEnabled"].(bool)
	if !enabled {
		return cfg, "", nil, false
	}
	modelID, _ := data["summaryModelId"].(string)
	var facts []string
	if raw, ok := data["userMemories"].([]any); ok {
		for _, v := range raw {
			if s, ok := v.(string); ok {
				facts = append(facts, s)
			}
		}
	}
	return cfg, modelID, facts, true
}

func (e *Extractor) saveUserFacts(ctx context.Context, cfg types.Config, facts []string) {
	var data map[string]any
	_ = json.Unmarshal(cfg.Data, &data)
	if data == nil {
		data = map[string]any{}
	}
	data["userMemories"] = facts
	cfg.Data, _ = json.Marshal(data)
	if _, err := e.configStore.Update(ctx, []types.Config{cfg}); err != nil {
		log.Printf("memory: save user facts: %v", err)
	}
}

func sanitizeName(name string) string {
	r := strings.NewReplacer(" ", "_", "-", "_", ".", "_")
	return strings.ToLower(r.Replace(name))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
