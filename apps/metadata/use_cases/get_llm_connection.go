package usecases

import (
	"context"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type GetLlmConnection struct {
	store protocols.Store[string, types.LlmConnection]
}

func NewGetLlmConnection(store protocols.Store[string, types.LlmConnection]) *GetLlmConnection {
	return &GetLlmConnection{store: store}
}

func (uc *GetLlmConnection) Execute(ctx context.Context, id string) (types.LlmConnection, error) {
	result, err := uc.store.Get(ctx, []string{id})
	if err != nil {
		return types.LlmConnection{}, err
	}
	c, ok := result[id]
	if !ok {
		return types.LlmConnection{}, base.ErrNotFound
	}
	return c, nil
}
