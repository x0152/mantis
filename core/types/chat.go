package types

import (
	"encoding/json"
	"time"
)

type ChatSession struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Source    string    `json:"source,omitempty"`
	Active    bool      `json:"active,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type Attachment struct {
	ID       string `json:"id"`
	FileName string `json:"fileName"`
	MimeType string `json:"mimeType"`
	Size     int64  `json:"size"`
}

type ChatMessage struct {
	ID          string          `json:"id"`
	SessionID   string          `json:"sessionId"`
	Role        string          `json:"role"`
	Content     string          `json:"content"`
	Status      string          `json:"status"`
	Source      string          `json:"source,omitempty"`
	ModelID     string          `json:"modelId,omitempty"`
	ModelName   string          `json:"modelName,omitempty"`
	PresetID    string          `json:"presetId,omitempty"`
	PresetName  string          `json:"presetName,omitempty"`
	ModelRole   string          `json:"modelRole,omitempty"` // primary | fallback | explicit | legacy
	Steps       json.RawMessage `json:"steps,omitempty"`
	Attachments []Attachment    `json:"attachments,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
}
