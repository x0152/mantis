package models

import "github.com/uptrace/bun"

type ModelRow struct {
	bun.BaseModel `bun:"table:models"`
	ID            string `bun:"id,pk"`
	ConnectionID  string `bun:"connection_id"`
	Name          string `bun:"name"`
	ThinkingMode  string `bun:"thinking_mode"`
}
