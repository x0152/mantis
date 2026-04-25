package usecases

import (
	"context"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type UpdateModel struct {
	store protocols.Store[string, types.Model]
}

func NewUpdateModel(store protocols.Store[string, types.Model]) *UpdateModel {
	return &UpdateModel{store: store}
}

func (uc *UpdateModel) Execute(ctx context.Context, m types.Model) (types.Model, error) {
	existing, err := uc.store.Get(ctx, []string{m.ID})
	if err != nil {
		return types.Model{}, err
	}
	if _, ok := existing[m.ID]; !ok {
		return types.Model{}, base.ErrNotFound
	}
	applyModelDefaults(&m)
	result, err := uc.store.Update(ctx, []types.Model{m})
	if err != nil {
		return types.Model{}, err
	}
	return result[0], nil
}
