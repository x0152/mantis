package usecases

import (
	"context"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type UpdateChannel struct {
	store protocols.Store[string, types.Channel]
}

func NewUpdateChannel(store protocols.Store[string, types.Channel]) *UpdateChannel {
	return &UpdateChannel{store: store}
}

func (uc *UpdateChannel) Execute(ctx context.Context, id, name, token, modelID string, allowedUserIDs []int64) (types.Channel, error) {
	existing, err := uc.store.Get(ctx, []string{id})
	if err != nil {
		return types.Channel{}, err
	}
	old, ok := existing[id]
	if !ok {
		return types.Channel{}, base.ErrNotFound
	}
	c := old
	c.Name = name
	c.Token = token
	c.ModelID = modelID
	if allowedUserIDs == nil {
		allowedUserIDs = []int64{}
	}
	c.AllowedUserIDs = allowedUserIDs
	if c.Type == "chat" {
		c.Token = ""
		c.AllowedUserIDs = []int64{}
	}
	result, err := uc.store.Update(ctx, []types.Channel{c})
	if err != nil {
		return types.Channel{}, err
	}
	return result[0], nil
}
