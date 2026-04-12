package mappers

import (
	"mantis/core/types"
	"mantis/infrastructure/models"
)

func PresetToRow(p types.Preset) models.PresetRow {
	return models.PresetRow{
		ID:              p.ID,
		Name:            p.Name,
		ChatModelID:     p.ChatModelID,
		SummaryModelID:  p.SummaryModelID,
		ImageModelID:    p.ImageModelID,
		FallbackModelID: p.FallbackModelID,
		Temperature:     p.Temperature,
		SystemPrompt:    p.SystemPrompt,
	}
}

func PresetFromRow(r models.PresetRow) types.Preset {
	return types.Preset{
		ID:              r.ID,
		Name:            r.Name,
		ChatModelID:     r.ChatModelID,
		SummaryModelID:  r.SummaryModelID,
		ImageModelID:    r.ImageModelID,
		FallbackModelID: r.FallbackModelID,
		Temperature:     r.Temperature,
		SystemPrompt:    r.SystemPrompt,
	}
}
