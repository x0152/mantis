package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type DeleteGuardRule struct {
	store protocols.Store[string, types.GuardRule]
}

func NewDeleteGuardRule(store protocols.Store[string, types.GuardRule]) *DeleteGuardRule {
	return &DeleteGuardRule{store: store}
}

func (uc *DeleteGuardRule) Execute(ctx context.Context, id string) error {
	return uc.store.Delete(ctx, []string{id})
}
