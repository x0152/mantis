package models

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type PlanRunRow struct {
	bun.BaseModel `bun:"table:plan_runs"`
	ID            string          `bun:"id,pk"`
	PlanID        string          `bun:"plan_id"`
	Status        string          `bun:"status"`
	Trigger       string          `bun:"trigger"`
	Steps         json.RawMessage `bun:"steps,type:jsonb"`
	StartedAt     time.Time       `bun:"started_at"`
	FinishedAt    *time.Time      `bun:"finished_at"`
}
