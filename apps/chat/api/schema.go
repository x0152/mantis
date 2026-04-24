package api

import "mantis/core/types"

type SessionOutput struct {
	Body types.ChatSession
}

type SessionsOutput struct {
	Body []types.ChatSession
}

type ListSessionsInput struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

type CreateSessionInput struct {
	Body struct {
		Title string `json:"title"`
	}
}

type UpdateSessionInput struct {
	ID   string `path:"id"`
	Body struct {
		Title string `json:"title"`
	}
}

type DeleteSessionInput struct {
	ID string `path:"id"`
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

type FileAttachmentInput struct {
	FileName   string `json:"fileName" required:"true" minLength:"1"`
	MimeType   string `json:"mimeType,omitempty"`
	DataBase64 string `json:"dataBase64" required:"true" minLength:"1"`
	Caption    string `json:"caption,omitempty"`
}

type SendMessageInput struct {
	Body struct {
		SessionID string                `json:"sessionId" required:"true" minLength:"1"`
		Content   string                `json:"content"`
		Files     []FileAttachmentInput `json:"files,omitempty"`
	}
}

type SendMessageResponse struct {
	UserMessage      types.ChatMessage `json:"userMessage"`
	AssistantMessage types.ChatMessage `json:"assistantMessage"`
}

type SendMessageOutput struct {
	Body SendMessageResponse
}

type RegenerateInput struct {
	ID string `path:"id"`
}

type RegenerateResponse struct {
	AssistantMessage types.ChatMessage `json:"assistantMessage"`
}

type RegenerateOutput struct {
	Body RegenerateResponse
}

type StopSessionInput struct {
	ID string `path:"id"`
}

type StopSessionResponse struct {
	Stopped bool `json:"stopped"`
}

type StopSessionOutput struct {
	Body StopSessionResponse
}
