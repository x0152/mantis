package usecases

import (
	"context"
	"fmt"
	"strings"

	modelplugin "mantis/core/plugins/model"
	"mantis/core/protocols"
	"mantis/core/types"
	messageworkflow "mantis/core/workflows/message"
	"mantis/shared"
)

type RegenerateLast struct {
	workflow     *messageworkflow.Workflow
	messageStore protocols.Store[string, types.ChatMessage]
	limits       shared.Limits
}

func NewRegenerateLast(
	workflow *messageworkflow.Workflow,
	messageStore protocols.Store[string, types.ChatMessage],
	limits shared.Limits,
) *RegenerateLast {
	return &RegenerateLast{workflow: workflow, messageStore: messageStore, limits: limits}
}

func (uc *RegenerateLast) Execute(ctx context.Context, sessionID string) (types.ChatMessage, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return types.ChatMessage{}, fmt.Errorf("session_id is required")
	}

	messages, err := uc.messageStore.List(ctx, types.ListQuery{
		Filter: map[string]string{"session_id": sessionID},
		Sort:   []types.Sort{{Field: "created_at", Dir: types.SortDirDesc}},
		Page:   types.Page{Limit: 20},
	})
	if err != nil {
		return types.ChatMessage{}, err
	}
	if len(messages) == 0 {
		return types.ChatMessage{}, fmt.Errorf("no messages to regenerate")
	}

	var lastAssistant *types.ChatMessage
	for i := range messages {
		m := messages[i]
		if m.Role == "assistant" {
			lastAssistant = &m
			break
		}
	}
	if lastAssistant == nil {
		return types.ChatMessage{}, fmt.Errorf("no assistant message to regenerate")
	}
	if lastAssistant.Status == "pending" {
		return types.ChatMessage{}, fmt.Errorf("assistant message is still generating; stop it first")
	}

	var userMsg *types.ChatMessage
	for i := range messages {
		m := messages[i]
		if m.Role != "user" {
			continue
		}
		if !m.CreatedAt.Before(lastAssistant.CreatedAt) {
			continue
		}
		if userMsg == nil || m.CreatedAt.After(userMsg.CreatedAt) {
			userMsg = &m
		}
	}
	if userMsg == nil {
		return types.ChatMessage{}, fmt.Errorf("no preceding user message to regenerate from")
	}

	if err := uc.messageStore.Delete(ctx, []string{lastAssistant.ID}); err != nil {
		return types.ChatMessage{}, fmt.Errorf("delete previous assistant: %w", err)
	}

	source := strings.TrimSpace(lastAssistant.Source)
	if source == "" {
		source = "web"
	}

	return uc.workflow.Regenerate(ctx, messageworkflow.RegenerateInput{
		SessionID:   sessionID,
		UserContent: userMsg.Content,
		Source:      source,
		ModelConfig: modelplugin.Input{ChannelID: "chat", DefaultPreset: "chat"},
		Timeout:     uc.limits.SupervisorTimeout,
	})
}
