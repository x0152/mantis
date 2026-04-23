package mappers

import (
	"encoding/json"

	"mantis/core/types"
	"mantis/infrastructure/models"
)

func ChatSessionToRow(s types.ChatSession) models.ChatSessionRow {
	return models.ChatSessionRow{ID: s.ID, Title: s.Title, Source: s.Source, CreatedAt: s.CreatedAt}
}

func ChatSessionFromRow(r models.ChatSessionRow) types.ChatSession {
	return types.ChatSession{ID: r.ID, Title: r.Title, Source: r.Source, CreatedAt: r.CreatedAt}
}

func ChatMessageToRow(m types.ChatMessage) models.ChatMessageRow {
	var att json.RawMessage
	if len(m.Attachments) > 0 {
		att, _ = json.Marshal(m.Attachments)
	}
	return models.ChatMessageRow{
		ID: m.ID, SessionID: m.SessionID, Role: m.Role,
		Content: m.Content, Status: m.Status, Source: m.Source,
		ModelID: m.ModelID, ModelName: m.ModelName,
		PresetID: m.PresetID, PresetName: m.PresetName, ModelRole: m.ModelRole,
		Steps: m.Steps, Attachments: att, CreatedAt: m.CreatedAt, FinishedAt: m.FinishedAt,
	}
}

func ChatMessageFromRow(r models.ChatMessageRow) types.ChatMessage {
	var att []types.Attachment
	if len(r.Attachments) > 0 {
		_ = json.Unmarshal(r.Attachments, &att)
	}
	return types.ChatMessage{
		ID: r.ID, SessionID: r.SessionID, Role: r.Role,
		Content: r.Content, Status: r.Status, Source: r.Source,
		ModelID: r.ModelID, ModelName: r.ModelName,
		PresetID: r.PresetID, PresetName: r.PresetName, ModelRole: r.ModelRole,
		Steps: r.Steps, Attachments: att, CreatedAt: r.CreatedAt, FinishedAt: r.FinishedAt,
	}
}
