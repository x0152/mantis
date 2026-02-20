package mappers

import (
	"mantis/core/types"
	"mantis/infrastructure/models"
)

func ChatSessionToRow(s types.ChatSession) models.ChatSessionRow {
	return models.ChatSessionRow{ID: s.ID, CreatedAt: s.CreatedAt}
}

func ChatSessionFromRow(r models.ChatSessionRow) types.ChatSession {
	return types.ChatSession{ID: r.ID, CreatedAt: r.CreatedAt}
}

func ChatMessageToRow(m types.ChatMessage) models.ChatMessageRow {
	return models.ChatMessageRow{
		ID: m.ID, SessionID: m.SessionID, Role: m.Role,
		Content: m.Content, Status: m.Status, Source: m.Source,
		ModelName: m.ModelName, Steps: m.Steps, CreatedAt: m.CreatedAt,
	}
}

func ChatMessageFromRow(r models.ChatMessageRow) types.ChatMessage {
	return types.ChatMessage{
		ID: r.ID, SessionID: r.SessionID, Role: r.Role,
		Content: r.Content, Status: r.Status, Source: r.Source,
		ModelName: r.ModelName, Steps: r.Steps, CreatedAt: r.CreatedAt,
	}
}
