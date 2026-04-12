package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type ListPlans struct {
	store protocols.Store[string, types.Plan]
}

func NewListPlans(store protocols.Store[string, types.Plan]) *ListPlans {
	return &ListPlans{store: store}
}

func (uc *ListPlans) Execute(ctx context.Context) ([]types.Plan, error) {
	items, err := uc.store.List(ctx, types.ListQuery{})
	if items == nil {
		items = []types.Plan{}
	}
	return items, err
}
