package cron

import (
	"context"
	"log"

	usecases "mantis/apps/cron/use_cases"
	"mantis/core/types"
)

func (a *App) syncJobs(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.sched == nil || a.ucSyncJobs == nil {
		return
	}

	next, err := a.ucSyncJobs.Execute(ctx, usecases.SyncJobsInput{
		Scheduler: a.sched,
		Entries:   a.entries,
		Run: func(job types.CronJob) {
			a.executeJob(context.Background(), job)
		},
	})
	if err != nil {
		log.Printf("cron: list jobs: %v", err)
		return
	}
	a.entries = next
}
