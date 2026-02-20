package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type CreateLlmConnection struct {
	store protocols.Store[string, types.LlmConnection]
}

func NewCreateLlmConnection(store protocols.Store[string, types.LlmConnection]) *CreateLlmConnection {
	return &CreateLlmConnection{store: store}
}

func (uc *CreateLlmConnection) Execute(ctx context.Context, id, provider, baseURL, apiKey string) (types.LlmConnection, error) {
	c := types.LlmConnection{ID: id, Provider: provider, BaseURL: baseURL, APIKey: apiKey}
	result, err := uc.store.Create(ctx, []types.LlmConnection{c})
	if err != nil {
		return types.LlmConnection{}, err
	}
	return result[0], nil
}
