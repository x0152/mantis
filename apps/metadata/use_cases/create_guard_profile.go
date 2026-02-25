package usecases

import (
	"context"

	"github.com/google/uuid"

	"mantis/core/protocols"
	"mantis/core/types"
)

type CreateGuardProfile struct {
	store protocols.Store[string, types.GuardProfile]
}

func NewCreateGuardProfile(store protocols.Store[string, types.GuardProfile]) *CreateGuardProfile {
	return &CreateGuardProfile{store: store}
}

func (uc *CreateGuardProfile) Execute(ctx context.Context, name, description string, capabilities types.GuardCapabilities, commands []types.CommandRule) (types.GuardProfile, error) {
	p := types.GuardProfile{
		ID:           uuid.New().String(),
		Name:         name,
		Description:  description,
		Builtin:      false,
		Capabilities: capabilities,
		Commands:     commands,
	}
	if p.Commands == nil {
		p.Commands = []types.CommandRule{}
	}
	result, err := uc.store.Create(ctx, []types.GuardProfile{p})
	if err != nil {
		return types.GuardProfile{}, err
	}
	return result[0], nil
}
