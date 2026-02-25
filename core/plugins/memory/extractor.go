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

const userPrompt = `You manage a personal memory system about the user.
You receive existing facts and a chat conversation.

Return the COMPLETE updated list of facts about the user. You may add, modify, or remove.
- One concise line per fact.
- Only long-term preferences, habits, context. NOT transient requests.
- Remove outdated or contradicted facts. No duplicates.

Return strictly valid JSON: ["fact1", "fact2"]
Empty if nothing worth remembering: []`

const connectionPrompt = `You manage a memory system about a remote server.
You receive existing facts and SSH command history (tasks + outputs).

Return the COMPLETE updated list of facts about this server. You may add, modify, or remove.
- One concise line per fact (installed software, config details, services, issues found, etc).
- Only persistent/useful server state. NOT one-off command outputs.
- Remove outdated or contradicted facts. No duplicates.

Return strictly valid JSON: ["fact1", "fact2"]
Empty if nothing worth remembering: []`

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
	input := fmt.Sprintf("Existing facts:\n%s\n\nConversation:\nUser: %s\nAssistant: %s", string(existingJSON), userContent, assistantContent)

	result, err := e.callLLM(ctx, llmConn, model, userPrompt, input)
	if err != nil {
		log.Printf("memory: user extract: %v", err)
		return
	}

	var facts []string
	if err := json.Unmarshal([]byte(result), &facts); err != nil {
		log.Printf("memory: parse user facts: %v (raw: %s)", err, result)
		return
	}

	e.saveUserFacts(ctx, cfg, facts)
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

		input := fmt.Sprintf("Existing facts:\n%s\n\nSSH history:\n%s", string(existingJSON), history.String())

		result, err := e.callLLM(ctx, llmConn, model, connectionPrompt, input)
		if err != nil {
			log.Printf("memory: connection %s extract: %v", connID, err)
			continue
		}

		var facts []string
		if err := json.Unmarshal([]byte(result), &facts); err != nil {
			log.Printf("memory: parse connection %s facts: %v (raw: %s)", connID, err, result)
			continue
		}

		now := time.Now()
		memories := make([]types.Memory, len(facts))
		for i, f := range facts {
			memories[i] = types.Memory{
				ID:        fmt.Sprintf("%s-%d", connID[:8], i),
				Content:   f,
				CreatedAt: now,
			}
		}
		c.Memories = memories
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
	if idx := strings.Index(raw, "["); idx >= 0 {
		raw = raw[idx:]
	}
	if end := strings.LastIndex(raw, "]"); end >= 0 {
		raw = raw[:end+1]
	}
	return raw, nil
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
