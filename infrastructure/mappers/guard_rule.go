package mappers

import (
	"mantis/core/types"
	"mantis/infrastructure/models"
)

func GuardRuleToRow(r types.GuardRule) models.GuardRuleRow {
	return models.GuardRuleRow{
		ID: r.ID, Name: r.Name, Description: r.Description,
		Pattern: r.Pattern, ConnectionID: r.ConnectionID, Enabled: r.Enabled,
	}
}

func GuardRuleFromRow(r models.GuardRuleRow) types.GuardRule {
	return types.GuardRule{
		ID: r.ID, Name: r.Name, Description: r.Description,
		Pattern: r.Pattern, ConnectionID: r.ConnectionID, Enabled: r.Enabled,
	}
}
