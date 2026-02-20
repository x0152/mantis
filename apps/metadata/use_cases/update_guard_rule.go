package usecases

import (
	"context"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type UpdateGuardRule struct {
	store protocols.Store[string, types.GuardRule]
}

func NewUpdateGuardRule(store protocols.Store[string, types.GuardRule]) *UpdateGuardRule {
	return &UpdateGuardRule{store: store}
}

func (uc *UpdateGuardRule) Execute(ctx context.Context, id, name, description, pattern, connectionID string, enabled bool) (types.GuardRule, error) {
	existing, err := uc.store.Get(ctx, []string{id})
	if err != nil {
		return types.GuardRule{}, err
	}
	if _, ok := existing[id]; !ok {
		return types.GuardRule{}, base.ErrNotFound
	}
	r := types.GuardRule{
		ID: id, Name: name, Description: description,
		Pattern: pattern, ConnectionID: connectionID, Enabled: enabled,
	}
	result, err := uc.store.Update(ctx, []types.GuardRule{r})
	if err != nil {
		return types.GuardRule{}, err
	}
	return result[0], nil
}
