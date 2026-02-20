package models

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type ChatSessionRow struct {
	bun.BaseModel `bun:"table:chat_sessions"`
	ID            string    `bun:"id,pk"`
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
	ModelName     string          `bun:"model_name"`
	Steps         json.RawMessage `bun:"steps,type:jsonb"`
	CreatedAt     time.Time       `bun:"created_at"`
}
