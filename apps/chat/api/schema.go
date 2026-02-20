package api

import "mantis/core/types"

type SessionOutput struct {
	Body types.ChatSession
}

type MessagesOutput struct {
	Body []types.ChatMessage
}

type ListMessagesInput struct {
	Limit     int    `query:"limit"`
	Offset    int    `query:"offset"`
	SessionID string `query:"sessionId"`
	Source    string `query:"source"`
}

type SendMessageInput struct {
	Body struct {
		SessionID string `json:"sessionId" required:"true" minLength:"1"`
		Content   string `json:"content" required:"true" minLength:"1"`
	}
}

type SendMessageResponse struct {
	UserMessage      types.ChatMessage `json:"userMessage"`
	AssistantMessage types.ChatMessage `json:"assistantMessage"`
}

type SendMessageOutput struct {
	Body SendMessageResponse
}
