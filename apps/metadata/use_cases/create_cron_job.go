package usecases

import (
	"context"

	"github.com/google/uuid"

	"mantis/core/protocols"
	"mantis/core/types"
)

type CreateCronJob struct {
	store protocols.Store[string, types.CronJob]
}

func NewCreateCronJob(store protocols.Store[string, types.CronJob]) *CreateCronJob {
	return &CreateCronJob{store: store}
}

func (uc *CreateCronJob) Execute(ctx context.Context, name, schedule, prompt string, enabled bool) (types.CronJob, error) {
	j := types.CronJob{
		ID:       uuid.New().String(),
		Name:     name,
		Schedule: schedule,
		Prompt:   prompt,
		Enabled:  enabled,
	}
	result, err := uc.store.Create(ctx, []types.CronJob{j})
	if err != nil {
		return types.CronJob{}, err
	}
	return result[0], nil
}
