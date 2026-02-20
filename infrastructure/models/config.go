package models

import (
	"encoding/json"

	"github.com/uptrace/bun"
)

type ConfigRow struct {
	bun.BaseModel `bun:"table:configs"`
	ID            string          `bun:"id,pk"`
	Data          json.RawMessage `bun:"data,type:jsonb"`
}
