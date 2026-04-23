package models

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type ChatSessionRow struct {
	bun.BaseModel `bun:"table:chat_sessions"`
	ID            string    `bun:"id,pk"`
	Title         string    `bun:"title"`
	Source        string    `bun:"source"`
	CreatedAt     time.Time `bun:"created_at"`
}

type ChatMessageRow struct {
	bun.BaseModel `bun:"table:chat_messages"`
	ID            string          `bun:"id,pk"`
	SessionID     string          `bun:"session_id"`
	Role          string          `bun:"role"`
	Content       string          `bun:"content"`
	Status        string          `bun:"status"`
	Source        string          `bun:"source"`
	ModelID       string          `bun:"model_id"`
	ModelName     string          `bun:"model_name"`
	PresetID      string          `bun:"preset_id"`
	PresetName    string          `bun:"preset_name"`
	ModelRole     string          `bun:"model_role"`
	Steps         json.RawMessage `bun:"steps,type:jsonb"`
	Attachments   json.RawMessage `bun:"attachments,type:jsonb"`
	CreatedAt     time.Time       `bun:"created_at"`
	FinishedAt    *time.Time      `bun:"finished_at"`
}
