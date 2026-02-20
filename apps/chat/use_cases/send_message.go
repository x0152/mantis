package usecases

import (
	"context"
	"time"

	modelplugin "mantis/core/plugins/model"
	"mantis/core/types"
	messageworkflow "mantis/core/workflows/message"
)

type SendMessage struct {
	workflow *messageworkflow.Workflow
}

func NewSendMessage(workflow *messageworkflow.Workflow) *SendMessage {
	return &SendMessage{workflow: workflow}
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
	return out.UserMessage, out.AssistantMessage, nil
}
