package models

import (
	"encoding/json"

	"github.com/uptrace/bun"
)

type SkillRow struct {
	bun.BaseModel `bun:"table:skills"`
	ID            string          `bun:"id,pk"`
	ConnectionID  string          `bun:"connection_id"`
	Name          string          `bun:"name"`
	Description   string          `bun:"description"`
	Parameters    json.RawMessage `bun:"parameters,type:jsonb"`
	Script        string          `bun:"script"`
}
