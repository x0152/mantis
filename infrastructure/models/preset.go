package models

import "github.com/uptrace/bun"

type PresetRow struct {
	bun.BaseModel   `bun:"table:presets"`
	ID              string   `bun:"id,pk"`
	Name            string   `bun:"name"`
	ChatModelID     string   `bun:"chat_model_id"`
	SummaryModelID  string   `bun:"summary_model_id"`
	ImageModelID    string   `bun:"image_model_id"`
	FallbackModelID string   `bun:"fallback_model_id"`
	Temperature     *float64 `bun:"temperature"`
	SystemPrompt    string   `bun:"system_prompt"`
}
