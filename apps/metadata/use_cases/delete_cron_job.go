package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type DeleteCronJob struct {
	store protocols.Store[string, types.CronJob]
}

func NewDeleteCronJob(store protocols.Store[string, types.CronJob]) *DeleteCronJob {
	return &DeleteCronJob{store: store}
}

func (uc *DeleteCronJob) Execute(ctx context.Context, id string) error {
	return uc.store.Delete(ctx, []string{id})
}
