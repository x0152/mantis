package models

import (
	"encoding/json"

	"github.com/uptrace/bun"
)

type SettingsRow struct {
	bun.BaseModel  `bun:"table:settings"`
	ID             string          `bun:"id,pk"`
	ChatPresetID   string          `bun:"chat_preset_id"`
	ServerPresetID string          `bun:"server_preset_id"`
	MemoryEnabled  bool            `bun:"memory_enabled"`
	UserMemories   json.RawMessage `bun:"user_memories,type:jsonb"`
}
