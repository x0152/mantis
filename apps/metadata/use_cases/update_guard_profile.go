package usecases

import (
	"context"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type UpdateGuardProfile struct {
	store protocols.Store[string, types.GuardProfile]
}

func NewUpdateGuardProfile(store protocols.Store[string, types.GuardProfile]) *UpdateGuardProfile {
	return &UpdateGuardProfile{store: store}
}

func (uc *UpdateGuardProfile) Execute(ctx context.Context, id, name, description string, capabilities types.GuardCapabilities, commands []types.CommandRule) (types.GuardProfile, error) {
	existing, err := uc.store.Get(ctx, []string{id})
	if err != nil {
		return types.GuardProfile{}, err
	}
	old, ok := existing[id]
	if !ok {
		return types.GuardProfile{}, base.ErrNotFound
	}
	if commands == nil {
		commands = []types.CommandRule{}
	}
	p := types.GuardProfile{
		ID:           id,
		Name:         name,
		Description:  description,
		Builtin:      old.Builtin,
		Capabilities: capabilities,
		Commands:     commands,
	}
	result, err := uc.store.Update(ctx, []types.GuardProfile{p})
	if err != nil {
		return types.GuardProfile{}, err
	}
	return result[0], nil
}
