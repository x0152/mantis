package usecases

import (
	"context"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type GetChannel struct {
	store protocols.Store[string, types.Channel]
}

func NewGetChannel(store protocols.Store[string, types.Channel]) *GetChannel {
	return &GetChannel{store: store}
}

func (uc *GetChannel) Execute(ctx context.Context, id string) (types.Channel, error) {
	result, err := uc.store.Get(ctx, []string{id})
	if err != nil {
		return types.Channel{}, err
	}
	c, ok := result[id]
	if !ok {
		return types.Channel{}, base.ErrNotFound
	}
	return c, nil
}
