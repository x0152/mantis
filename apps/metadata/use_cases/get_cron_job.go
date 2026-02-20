package usecases

import (
	"context"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type GetCronJob struct {
	store protocols.Store[string, types.CronJob]
}

func NewGetCronJob(store protocols.Store[string, types.CronJob]) *GetCronJob {
	return &GetCronJob{store: store}
}

func (uc *GetCronJob) Execute(ctx context.Context, id string) (types.CronJob, error) {
	result, err := uc.store.Get(ctx, []string{id})
	if err != nil {
		return types.CronJob{}, err
	}
	j, ok := result[id]
	if !ok {
		return types.CronJob{}, base.ErrNotFound
	}
	return j, nil
}
