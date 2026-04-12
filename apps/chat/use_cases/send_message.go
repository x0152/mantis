package usecases

import (
	"context"
	"time"
	"unicode/utf8"

	modelplugin "mantis/core/plugins/model"
	"mantis/core/protocols"
	"mantis/core/types"
	messageworkflow "mantis/core/workflows/message"
)

type SendMessage struct {
	workflow     *messageworkflow.Workflow
	sessionStore protocols.Store[string, types.ChatSession]
}

func NewSendMessage(workflow *messageworkflow.Workflow, sessionStore protocols.Store[string, types.ChatSession]) *SendMessage {
	return &SendMessage{workflow: workflow, sessionStore: sessionStore}
}

func (uc *SendMessage) Execute(ctx context.Context, sessionID, content string) (types.ChatMessage, types.ChatMessage, error) {
	out, err := uc.workflow.Execute(ctx, messageworkflow.Input{
		SessionID:   sessionID,
		Content:     content,
		Source:      "web",
		ModelConfig: modelplugin.Input{ChannelID: "chat", DefaultPreset: "chat"},
		Timeout:     5 * time.Minute,
	})
	if err != nil {
		return types.ChatMessage{}, types.ChatMessage{}, err
	}

	go uc.autoTitle(context.Background(), sessionID, content)

	return out.UserMessage, out.AssistantMessage, nil
}

func (uc *SendMessage) autoTitle(ctx context.Context, sessionID, content string) {
	sessions, err := uc.sessionStore.Get(ctx, []string{sessionID})
	if err != nil {
		return
	}
	session, ok := sessions[sessionID]
	if !ok || session.Title != "" {
		return
	}

	title := content
	if utf8.RuneCountInString(title) > 50 {
		runes := []rune(title)
		title = string(runes[:50]) + "..."
	}
	session.Title = title
	_, _ = uc.sessionStore.Update(ctx, []types.ChatSession{session})
}
