package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type ListConnections struct {
	store protocols.Store[string, types.Connection]
}

func NewListConnections(store protocols.Store[string, types.Connection]) *ListConnections {
	return &ListConnections{store: store}
}

func (uc *ListConnections) Execute(ctx context.Context) ([]types.Connection, error) {
	items, err := uc.store.List(ctx, types.ListQuery{})
	if items == nil {
		items = []types.Connection{}
	}
	return items, err
}
