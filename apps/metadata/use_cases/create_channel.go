package usecases

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type CreateChannel struct {
	store protocols.Store[string, types.Channel]
}

func NewCreateChannel(store protocols.Store[string, types.Channel]) *CreateChannel {
	return &CreateChannel{store: store}
}

func (uc *CreateChannel) Execute(ctx context.Context, chType, name, token, modelID string, allowedUserIDs []int64) (types.Channel, error) {
	if strings.TrimSpace(chType) != "telegram" {
		return types.Channel{}, base.ErrValidation
	}
	if strings.TrimSpace(token) == "" {
		return types.Channel{}, base.ErrValidation
	}
	if allowedUserIDs == nil {
		allowedUserIDs = []int64{}
	}
	c := types.Channel{
		ID:             uuid.New().String(),
		Type:           "telegram",
		Name:           name,
		Token:          token,
		ModelID:        modelID,
		AllowedUserIDs: allowedUserIDs,
	}
	result, err := uc.store.Create(ctx, []types.Channel{c})
	if err != nil {
		return types.Channel{}, err
	}
	return result[0], nil
}
