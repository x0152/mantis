package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type ListModels struct {
	store protocols.Store[string, types.Model]
}

func NewListModels(store protocols.Store[string, types.Model]) *ListModels {
	return &ListModels{store: store}
}

func (uc *ListModels) Execute(ctx context.Context) ([]types.Model, error) {
	items, err := uc.store.List(ctx, types.ListQuery{})
	if items == nil {
		items = []types.Model{}
	}
	return items, err
}
