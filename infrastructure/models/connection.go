package models

import (
	"encoding/json"

	"github.com/uptrace/bun"
)

type ConnectionRow struct {
	bun.BaseModel `bun:"table:connections"`
	ID            string          `bun:"id,pk"`
	Type          string          `bun:"type"`
	Name          string          `bun:"name"`
	Description   string          `bun:"description"`
	ModelID       string          `bun:"model_id"`
	Config        json.RawMessage `bun:"config,type:jsonb"`
	Memories      json.RawMessage `bun:"memories,type:jsonb"`
	ProfileIDs    json.RawMessage `bun:"profile_ids,type:jsonb"`
	MemoryEnabled bool            `bun:"memory_enabled"`
}
