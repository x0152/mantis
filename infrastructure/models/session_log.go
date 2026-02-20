package models

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type SessionLogRow struct {
	bun.BaseModel `bun:"table:session_logs"`
	ID            string          `bun:"id,pk"`
	ConnectionID  string          `bun:"connection_id"`
	AgentName     string          `bun:"agent_name"`
	Prompt        string          `bun:"prompt"`
	Status        string          `bun:"status"`
	MessageID     string          `bun:"message_id"`
	StepID        string          `bun:"step_id"`
	ModelName     string          `bun:"model_name"`
	Entries       json.RawMessage `bun:"entries,type:jsonb"`
	StartedAt     time.Time       `bun:"started_at"`
	FinishedAt    *time.Time      `bun:"finished_at"`
}
