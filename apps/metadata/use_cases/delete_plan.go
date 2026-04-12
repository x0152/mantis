package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type DeletePlan struct {
	store protocols.Store[string, types.Plan]
}

func NewDeletePlan(store protocols.Store[string, types.Plan]) *DeletePlan {
	return &DeletePlan{store: store}
}

func (uc *DeletePlan) Execute(ctx context.Context, id string) error {
	return uc.store.Delete(ctx, []string{id})
}
