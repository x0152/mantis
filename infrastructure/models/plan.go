package models

import (
	"encoding/json"

	"github.com/uptrace/bun"
)

type PlanRow struct {
	bun.BaseModel `bun:"table:plans"`
	ID            string          `bun:"id,pk"`
	Name          string          `bun:"name"`
	Description   string          `bun:"description"`
	Schedule      string          `bun:"schedule"`
	Enabled       bool            `bun:"enabled"`
	Graph         json.RawMessage `bun:"graph,type:jsonb"`
}
