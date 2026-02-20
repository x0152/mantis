package mappers

import (
	"mantis/core/types"
	"mantis/infrastructure/models"
)

func ConfigToRow(c types.Config) models.ConfigRow {
	return models.ConfigRow{ID: c.ID, Data: c.Data}
}

func ConfigFromRow(r models.ConfigRow) types.Config {
	return types.Config{ID: r.ID, Data: r.Data}
}
