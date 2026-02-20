package mappers

import (
	"mantis/core/types"
	"mantis/infrastructure/models"
)

func LlmConnectionToRow(c types.LlmConnection) models.LlmConnectionRow {
	return models.LlmConnectionRow{ID: c.ID, Provider: c.Provider, BaseURL: c.BaseURL, APIKey: c.APIKey}
}

func LlmConnectionFromRow(r models.LlmConnectionRow) types.LlmConnection {
	return types.LlmConnection{ID: r.ID, Provider: r.Provider, BaseURL: r.BaseURL, APIKey: r.APIKey}
}
