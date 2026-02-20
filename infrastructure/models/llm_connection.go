package models

import "github.com/uptrace/bun"

type LlmConnectionRow struct {
	bun.BaseModel `bun:"table:llm_connections"`
	ID            string `bun:"id,pk"`
	Provider      string `bun:"provider"`
	BaseURL       string `bun:"base_url"`
	APIKey        string `bun:"api_key"`
}
