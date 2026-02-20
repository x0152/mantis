package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type DeleteLlmConnection struct {
	store protocols.Store[string, types.LlmConnection]
}

func NewDeleteLlmConnection(store protocols.Store[string, types.LlmConnection]) *DeleteLlmConnection {
	return &DeleteLlmConnection{store: store}
}

func (uc *DeleteLlmConnection) Execute(ctx context.Context, id string) error {
	return uc.store.Delete(ctx, []string{id})
}
