package usecases

import (
	"context"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type GetPlanRun struct {
	store protocols.Store[string, types.PlanRun]
}

func NewGetPlanRun(store protocols.Store[string, types.PlanRun]) *GetPlanRun {
	return &GetPlanRun{store: store}
}

func (uc *GetPlanRun) Execute(ctx context.Context, id string) (types.PlanRun, error) {
	items, err := uc.store.Get(ctx, []string{id})
	if err != nil {
		return types.PlanRun{}, err
	}
	r, ok := items[id]
	if !ok {
		return types.PlanRun{}, base.ErrNotFound
	}
	return r, nil
}
