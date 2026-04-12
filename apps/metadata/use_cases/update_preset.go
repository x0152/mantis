package usecases

import (
	"context"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type UpdatePreset struct {
	store protocols.Store[string, types.Preset]
}

func NewUpdatePreset(store protocols.Store[string, types.Preset]) *UpdatePreset {
	return &UpdatePreset{store: store}
}

func (uc *UpdatePreset) Execute(ctx context.Context, p types.Preset) (types.Preset, error) {
	existing, err := uc.store.Get(ctx, []string{p.ID})
	if err != nil {
		return types.Preset{}, err
	}
	if _, ok := existing[p.ID]; !ok {
		return types.Preset{}, base.ErrNotFound
	}
	result, err := uc.store.Update(ctx, []types.Preset{p})
	if err != nil {
		return types.Preset{}, err
	}
	return result[0], nil
}
