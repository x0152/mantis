package models

import "github.com/uptrace/bun"

type GuardRuleRow struct {
	bun.BaseModel `bun:"table:guard_rules"`
	ID            string `bun:"id,pk"`
	Name          string `bun:"name"`
	Description   string `bun:"description"`
	Pattern       string `bun:"pattern"`
	ConnectionID  string `bun:"connection_id"`
	Enabled       bool   `bun:"enabled"`
}
