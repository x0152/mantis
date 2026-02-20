package usecases

import (
	"context"

	"github.com/google/uuid"

	"mantis/core/protocols"
	"mantis/core/types"
)

type CreateGuardRule struct {
	store protocols.Store[string, types.GuardRule]
}

func NewCreateGuardRule(store protocols.Store[string, types.GuardRule]) *CreateGuardRule {
	return &CreateGuardRule{store: store}
}

func (uc *CreateGuardRule) Execute(ctx context.Context, name, description, pattern, connectionID string, enabled bool) (types.GuardRule, error) {
	r := types.GuardRule{
		ID: uuid.New().String(), Name: name, Description: description,
		Pattern: pattern, ConnectionID: connectionID, Enabled: enabled,
	}
	result, err := uc.store.Create(ctx, []types.GuardRule{r})
	if err != nil {
		return types.GuardRule{}, err
	}
	return result[0], nil
}
