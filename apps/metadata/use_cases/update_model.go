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

func (uc *UpdateModel) Execute(ctx context.Context, id, connectionID, name, thinkingMode string) (types.Model, error) {
	existing, err := uc.store.Get(ctx, []string{id})
	if err != nil {
		return types.Model{}, err
	}
	if _, ok := existing[id]; !ok {
		return types.Model{}, base.ErrNotFound
	}
	m := types.Model{ID: id, ConnectionID: connectionID, Name: name, ThinkingMode: thinkingMode}
	result, err := uc.store.Update(ctx, []types.Model{m})
	if err != nil {
		return types.Model{}, err
	}
	return result[0], nil
}
