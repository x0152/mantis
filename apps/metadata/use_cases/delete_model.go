package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type DeleteModel struct {
	store protocols.Store[string, types.Model]
}

func NewDeleteModel(store protocols.Store[string, types.Model]) *DeleteModel {
	return &DeleteModel{store: store}
}

func (uc *DeleteModel) Execute(ctx context.Context, id string) error {
	return uc.store.Delete(ctx, []string{id})
}
