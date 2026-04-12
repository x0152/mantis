package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type DeletePreset struct {
	store protocols.Store[string, types.Preset]
}

func NewDeletePreset(store protocols.Store[string, types.Preset]) *DeletePreset {
	return &DeletePreset{store: store}
}

func (uc *DeletePreset) Execute(ctx context.Context, id string) error {
	return uc.store.Delete(ctx, []string{id})
}
