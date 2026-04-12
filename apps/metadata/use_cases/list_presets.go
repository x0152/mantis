package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type ListPresets struct {
	store protocols.Store[string, types.Preset]
}

func NewListPresets(store protocols.Store[string, types.Preset]) *ListPresets {
	return &ListPresets{store: store}
}

func (uc *ListPresets) Execute(ctx context.Context) ([]types.Preset, error) {
	items, err := uc.store.List(ctx, types.ListQuery{})
	if items == nil {
		items = []types.Preset{}
	}
	return items, err
}
