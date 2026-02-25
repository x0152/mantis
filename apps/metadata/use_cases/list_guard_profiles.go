package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type ListGuardProfiles struct {
	store protocols.Store[string, types.GuardProfile]
}

func NewListGuardProfiles(store protocols.Store[string, types.GuardProfile]) *ListGuardProfiles {
	return &ListGuardProfiles{store: store}
}

func (uc *ListGuardProfiles) Execute(ctx context.Context) ([]types.GuardProfile, error) {
	items, err := uc.store.List(ctx, types.ListQuery{})
	if items == nil {
		items = []types.GuardProfile{}
	}
	return items, err
}
