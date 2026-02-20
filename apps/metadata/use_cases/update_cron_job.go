package usecases

import (
	"context"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type UpdateCronJob struct {
	store protocols.Store[string, types.CronJob]
}

func NewUpdateCronJob(store protocols.Store[string, types.CronJob]) *UpdateCronJob {
	return &UpdateCronJob{store: store}
}

func (uc *UpdateCronJob) Execute(ctx context.Context, id, name, schedule, prompt string, enabled bool) (types.CronJob, error) {
	existing, err := uc.store.Get(ctx, []string{id})
	if err != nil {
		return types.CronJob{}, err
	}
	if _, ok := existing[id]; !ok {
		return types.CronJob{}, base.ErrNotFound
	}
	j := types.CronJob{
		ID:       id,
		Name:     name,
		Schedule: schedule,
		Prompt:   prompt,
		Enabled:  enabled,
	}
	result, err := uc.store.Update(ctx, []types.CronJob{j})
	if err != nil {
		return types.CronJob{}, err
	}
	return result[0], nil
}
