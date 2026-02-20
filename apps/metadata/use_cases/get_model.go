package usecases

import (
	"context"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type GetModel struct {
	store protocols.Store[string, types.Model]
}

func NewGetModel(store protocols.Store[string, types.Model]) *GetModel {
	return &GetModel{store: store}
}

func (uc *GetModel) Execute(ctx context.Context, id string) (types.Model, error) {
	result, err := uc.store.Get(ctx, []string{id})
	if err != nil {
		return types.Model{}, err
	}
	m, ok := result[id]
	if !ok {
		return types.Model{}, base.ErrNotFound
	}
	return m, nil
}
