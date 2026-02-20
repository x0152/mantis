package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type ListChannels struct {
	store protocols.Store[string, types.Channel]
}

func NewListChannels(store protocols.Store[string, types.Channel]) *ListChannels {
	return &ListChannels{store: store}
}

func (uc *ListChannels) Execute(ctx context.Context) ([]types.Channel, error) {
	items, err := uc.store.List(ctx, types.ListQuery{})
	if items == nil {
		items = []types.Channel{}
	}
	return items, err
}
