package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type DeleteConnection struct {
	store protocols.Store[string, types.Connection]
}

func NewDeleteConnection(store protocols.Store[string, types.Connection]) *DeleteConnection {
	return &DeleteConnection{store: store}
}

func (uc *DeleteConnection) Execute(ctx context.Context, id string) error {
	return uc.store.Delete(ctx, []string{id})
}
