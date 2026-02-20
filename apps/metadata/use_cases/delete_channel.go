package usecases

import (
	"context"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type DeleteChannel struct {
	store protocols.Store[string, types.Channel]
}

func NewDeleteChannel(store protocols.Store[string, types.Channel]) *DeleteChannel {
	return &DeleteChannel{store: store}
}

func (uc *DeleteChannel) Execute(ctx context.Context, id string) error {
	if id == "chat" {
		return base.ErrValidation
	}
	return uc.store.Delete(ctx, []string{id})
}
