package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type ListLlmConnections struct {
	store protocols.Store[string, types.LlmConnection]
}

func NewListLlmConnections(store protocols.Store[string, types.LlmConnection]) *ListLlmConnections {
	return &ListLlmConnections{store: store}
}

func (uc *ListLlmConnections) Execute(ctx context.Context) ([]types.LlmConnection, error) {
	items, err := uc.store.List(ctx, types.ListQuery{})
	if items == nil {
		items = []types.LlmConnection{}
	}
	return items, err
}
