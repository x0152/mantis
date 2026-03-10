package usecases

import (
	"context"
	"time"

	modelplugin "mantis/core/plugins/model"
	"mantis/core/protocols"
	"mantis/core/types"
	messageworkflow "mantis/core/workflows/message"
)

type SendMessage struct {
	workflow      *messageworkflow.Workflow
	sessionStore  protocols.Store[string, types.ChatSession]
	generateTitle *GenerateTitle
}

func NewSendMessage(
	workflow *messageworkflow.Workflow,
	sessionStore protocols.Store[string, types.ChatSession],
	generateTitle *GenerateTitle,
) *SendMessage {
	return &SendMessage{workflow: workflow, sessionStore: sessionStore, generateTitle: generateTitle}
}

func (uc *SendMessage) Execute(ctx context.Context, sessionID, content string) (types.ChatMessage, types.ChatMessage, error) {
	out, err := uc.workflow.Execute(ctx, messageworkflow.Input{
		SessionID:   sessionID,
		Content:     content,
		Source:      "web",
		ModelConfig: modelplugin.Input{ChannelID: "chat", ConfigPath: []string{"chat", "model_id"}},
		Timeout:     5 * time.Minute,
	})
	if err != nil {
		return types.ChatMessage{}, types.ChatMessage{}, err
	}

	go uc.generateTitle.Execute(context.Background(), sessionID, content)

	return out.UserMessage, out.AssistantMessage, nil
}
