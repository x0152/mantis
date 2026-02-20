package usecases

import (
	"context"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type GetConnection struct {
	store protocols.Store[string, types.Connection]
}

func NewGetConnection(store protocols.Store[string, types.Connection]) *GetConnection {
	return &GetConnection{store: store}
}

func (uc *GetConnection) Execute(ctx context.Context, id string) (types.Connection, error) {
	result, err := uc.store.Get(ctx, []string{id})
	if err != nil {
		return types.Connection{}, err
	}
	c, ok := result[id]
	if !ok {
		return types.Connection{}, base.ErrNotFound
	}
	return c, nil
}
