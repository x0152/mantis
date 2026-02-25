package usecases

import (
	"context"
	"testing"

	robcron "github.com/robfig/cron/v3"

	"mantis/core/types"
)

type cronJobStoreMock struct {
	jobs []types.CronJob
}

func (m *cronJobStoreMock) Create(_ context.Context, _ []types.CronJob) ([]types.CronJob, error) {
	return nil, nil
}

func (m *cronJobStoreMock) Get(_ context.Context, _ []string) (map[string]types.CronJob, error) {
	return map[string]types.CronJob{}, nil
}

func (m *cronJobStoreMock) List(_ context.Context, _ types.ListQuery) ([]types.CronJob, error) {
	return m.jobs, nil
}

func (m *cronJobStoreMock) Update(_ context.Context, items []types.CronJob) ([]types.CronJob, error) {
	return items, nil
}

func (m *cronJobStoreMock) Delete(_ context.Context, _ []string) error {
	return nil
}

func TestSyncJobs_ReplacesEntriesAndAddsOnlyEnabledValid(t *testing.T) {
	s := robcron.New()
	oldID, err := s.AddFunc("@every 1h", func() {})
	if err != nil {
		t.Fatal(err)
	}

	store := &cronJobStoreMock{
		jobs: []types.CronJob{
			{ID: "job-1", Schedule: "@every 30m", Enabled: true},
			{ID: "job-2", Schedule: "@every 10m", Enabled: false},
			{ID: "job-3", Schedule: "bad", Enabled: true},
		},
	}
	uc := NewSyncJobs(store)

	next, err := uc.Execute(context.Background(), SyncJobsInput{
		Scheduler: s,
		Entries:   map[string]robcron.EntryID{"old": oldID},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(next) != 1 {
		t.Fatalf("unexpected entries len: %d", len(next))
	}
	if _, ok := next["job-1"]; !ok {
		t.Fatalf("job-1 not scheduled: %+v", next)
	}

	entries := s.Entries()
	if len(entries) != 1 {
		t.Fatalf("unexpected scheduler entries len: %d", len(entries))
	}
	if entries[0].ID == oldID {
		t.Fatal("old entry was not removed")
	}
}
