package cron

import (
	"context"
	"log"
	"mantis/core/types"
)

func (a *App) executeJob(ctx context.Context, job types.CronJob) {
	if a.ucExecuteJob == nil {
		return
	}
	out, err := a.ucExecuteJob.Execute(ctx, job)
	if err != nil {
		log.Printf("cron: execute job id=%s: %v", job.ID, err)
		return
	}
	if out.Skipped {
		log.Printf("cron: job already running, skip id=%s", job.ID)
		return
	}
	log.Printf("cron: job launched id=%s", job.ID)
}
