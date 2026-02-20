package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type ListCronJobs struct {
	store protocols.Store[string, types.CronJob]
}

func NewListCronJobs(store protocols.Store[string, types.CronJob]) *ListCronJobs {
	return &ListCronJobs{store: store}
}

func (uc *ListCronJobs) Execute(ctx context.Context) ([]types.CronJob, error) {
	items, err := uc.store.List(ctx, types.ListQuery{})
	if items == nil {
		items = []types.CronJob{}
	}
	return items, err
}
