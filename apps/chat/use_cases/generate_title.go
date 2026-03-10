package usecases

import (
	"context"
	"log"
	"strings"
	"time"

	modelplugin "mantis/core/plugins/model"
	"mantis/core/protocols"
	"mantis/core/types"
	"mantis/shared"
)

const titleSystemPrompt = `Generate a short title (3-6 words) for a chat conversation based on the user's first message.
The title should capture the main topic or intent.
Reply with ONLY the title text, nothing else. No quotes, no punctuation at the end, no prefixes.
Use the same language as the user's message.`

type GenerateTitle struct {
	llm           protocols.LLM
	modelResolver *modelplugin.Resolver
	modelStore    protocols.Store[string, types.Model]
	llmConnStore  protocols.Store[string, types.LlmConnection]
	sessionStore  protocols.Store[string, types.ChatSession]
}

func NewGenerateTitle(
	llm protocols.LLM,
	modelResolver *modelplugin.Resolver,
	modelStore protocols.Store[string, types.Model],
	llmConnStore protocols.Store[string, types.LlmConnection],
	sessionStore protocols.Store[string, types.ChatSession],
) *GenerateTitle {
	return &GenerateTitle{
		llm:           llm,
		modelResolver: modelResolver,
		modelStore:    modelStore,
		llmConnStore:  llmConnStore,
		sessionStore:  sessionStore,
	}
}

func (uc *GenerateTitle) Execute(ctx context.Context, sessionID, userMessage string) {
	sessions, err := uc.sessionStore.Get(ctx, []string{sessionID})
	if err != nil {
		return
	}
	session, ok := sessions[sessionID]
	if !ok || session.Title != "" {
		return
	}

	title := uc.generate(ctx, userMessage)
	if title == "" {
		title = truncateContent(userMessage, 50)
	}

	session.Title = title
	if _, err := uc.sessionStore.Update(ctx, []types.ChatSession{session}); err != nil {
		log.Printf("generate_title: save: %v", err)
	}
}

func (uc *GenerateTitle) generate(ctx context.Context, userMessage string) string {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resolved, err := uc.modelResolver.Execute(ctx, modelplugin.Input{
		ChannelID:  "chat",
		ConfigPath: []string{"chat", "model_id"},
	})
	if err != nil || resolved.ModelID == "" {
		return ""
	}

	model, err := shared.ResolveModel(ctx, uc.modelStore, resolved.ModelID)
	if err != nil {
		return ""
	}
	llmConn, err := shared.ResolveConnection(ctx, uc.llmConnStore, model.ConnectionID)
	if err != nil {
		return ""
	}

	messages := []protocols.LLMMessage{
		{Role: "system", Content: titleSystemPrompt},
		{Role: "user", Content: userMessage},
	}

	stream, err := uc.llm.ChatStream(ctx, llmConn.BaseURL, llmConn.APIKey, messages, model.Name, nil, "skip")
	if err != nil {
		log.Printf("generate_title: llm call: %v", err)
		return ""
	}

	var sb strings.Builder
	for event := range stream {
		if event.Type == "text" {
			sb.WriteString(event.Delta)
		}
	}

	title := strings.TrimSpace(sb.String())
	if strings.Contains(title, "\n") {
		title = lastMeaningfulLine(title)
	}
	title = strings.Trim(title, "\"'")
	return truncateContent(title, 80)
}

func lastMeaningfulLine(s string) string {
	lines := strings.Split(s, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, "*") ||
			strings.HasPrefix(line, "-") ||
			strings.HasSuffix(line, ":") ||
			(len(line) > 2 && line[1] == '.' && line[0] >= '0' && line[0] <= '9') {
			continue
		}
		line = strings.Trim(line, "*_`")
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func truncateContent(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "..."
}
