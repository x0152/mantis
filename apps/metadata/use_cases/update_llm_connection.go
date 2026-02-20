package usecases

import (
	"context"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type UpdateLlmConnection struct {
	store protocols.Store[string, types.LlmConnection]
}

func NewUpdateLlmConnection(store protocols.Store[string, types.LlmConnection]) *UpdateLlmConnection {
	return &UpdateLlmConnection{store: store}
}

func (uc *UpdateLlmConnection) Execute(ctx context.Context, id, provider, baseURL, apiKey string) (types.LlmConnection, error) {
	existing, err := uc.store.Get(ctx, []string{id})
	if err != nil {
		return types.LlmConnection{}, err
	}
	if _, ok := existing[id]; !ok {
		return types.LlmConnection{}, base.ErrNotFound
	}
	c := types.LlmConnection{ID: id, Provider: provider, BaseURL: baseURL, APIKey: apiKey}
	result, err := uc.store.Update(ctx, []types.LlmConnection{c})
	if err != nil {
		return types.LlmConnection{}, err
	}
	return result[0], nil
}
