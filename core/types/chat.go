package types

import (
	"encoding/json"
	"time"
)

type ChatSession struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
}

type ChatMessage struct {
	ID        string          `json:"id"`
	SessionID string          `json:"sessionId"`
	Role      string          `json:"role"`
	Content   string          `json:"content"`
	Status    string          `json:"status"`
	Source    string          `json:"source,omitempty"`
	ModelName string          `json:"modelName,omitempty"`
	Steps     json.RawMessage `json:"steps,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
}
