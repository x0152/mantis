package models

import "github.com/uptrace/bun"

type CronJobRow struct {
	bun.BaseModel `bun:"table:cron_jobs"`
	ID            string `bun:"id,pk"`
	Name          string `bun:"name"`
	Schedule      string `bun:"schedule"`
	Prompt        string `bun:"prompt"`
	Enabled       bool   `bun:"enabled"`
}
