package mappers

import (
	"mantis/core/types"
	"mantis/infrastructure/models"
)

func ModelToRow(m types.Model) models.ModelRow {
	return models.ModelRow{
		ID:            m.ID,
		ConnectionID:  m.ConnectionID,
		Name:          m.Name,
		ThinkingMode:  m.ThinkingMode,
		ContextWindow: m.ContextWindow,
		ReserveTokens: m.ReserveTokens,
		CompactTokens: m.CompactTokens,
	}
}

func ModelFromRow(r models.ModelRow) types.Model {
	return types.Model{
		ID:            r.ID,
		ConnectionID:  r.ConnectionID,
		Name:          r.Name,
		ThinkingMode:  r.ThinkingMode,
		ContextWindow: r.ContextWindow,
		ReserveTokens: r.ReserveTokens,
		CompactTokens: r.CompactTokens,
	}
}
