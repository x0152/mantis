package mappers

import (
	"mantis/core/types"
	"mantis/infrastructure/models"
)

func CronJobToRow(c types.CronJob) models.CronJobRow {
	return models.CronJobRow{
		ID:       c.ID,
		Name:     c.Name,
		Schedule: c.Schedule,
		Prompt:   c.Prompt,
		Enabled:  c.Enabled,
	}
}

func CronJobFromRow(r models.CronJobRow) types.CronJob {
	return types.CronJob{
		ID:       r.ID,
		Name:     r.Name,
		Schedule: r.Schedule,
		Prompt:   r.Prompt,
		Enabled:  r.Enabled,
	}
}
