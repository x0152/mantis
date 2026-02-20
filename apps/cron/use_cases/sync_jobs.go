package usecases

import (
	"context"

	robcron "github.com/robfig/cron/v3"

	"mantis/core/protocols"
	"mantis/core/types"
)

type SyncJobsInput struct {
	Scheduler *robcron.Cron
	Entries   map[string]robcron.EntryID
	Run       func(job types.CronJob)
}

type SyncJobs struct {
	cronJobStore protocols.Store[string, types.CronJob]
}

func NewSyncJobs(cronJobStore protocols.Store[string, types.CronJob]) *SyncJobs {
	return &SyncJobs{cronJobStore: cronJobStore}
}

func (uc *SyncJobs) Execute(ctx context.Context, in SyncJobsInput) (map[string]robcron.EntryID, error) {
	if uc.cronJobStore == nil || in.Scheduler == nil {
		return in.Entries, nil
	}

	jobs, err := uc.cronJobStore.List(ctx, types.ListQuery{})
	if err != nil {
		return in.Entries, err
	}
	if jobs == nil {
		jobs = []types.CronJob{}
	}

	for _, eid := range in.Entries {
		in.Scheduler.Remove(eid)
	}
	next := make(map[string]robcron.EntryID, len(jobs))

	for _, j := range jobs {
		if !j.Enabled {
			continue
		}
		job := j
		eid, err := in.Scheduler.AddFunc(job.Schedule, func() {
			if in.Run != nil {
				in.Run(job)
			}
		})
		if err != nil {
			continue
		}
		next[job.ID] = eid
	}
	return next, nil
}
