package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type DeleteGuardProfile struct {
	store protocols.Store[string, types.GuardProfile]
}

func NewDeleteGuardProfile(store protocols.Store[string, types.GuardProfile]) *DeleteGuardProfile {
	return &DeleteGuardProfile{store: store}
}

func (uc *DeleteGuardProfile) Execute(ctx context.Context, id string) error {
	return uc.store.Delete(ctx, []string{id})
}
