package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type ListGuardRules struct {
	store protocols.Store[string, types.GuardRule]
}

func NewListGuardRules(store protocols.Store[string, types.GuardRule]) *ListGuardRules {
	return &ListGuardRules{store: store}
}

func (uc *ListGuardRules) Execute(ctx context.Context) ([]types.GuardRule, error) {
	items, err := uc.store.List(ctx, types.ListQuery{})
	if items == nil {
		items = []types.GuardRule{}
	}
	return items, err
}
