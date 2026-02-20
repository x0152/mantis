package types

import (
	"encoding/json"
	"time"
)

type Connection struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	ModelID     string          `json:"modelId"`
	Config      json.RawMessage `json:"config"`
	Memories    []Memory        `json:"memories"`
}

type Memory struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}
