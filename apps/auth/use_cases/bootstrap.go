package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"mantis/core/auth"
	"mantis/core/protocols"
	"mantis/core/types"
)

type Bootstrap struct {
	store protocols.Store[string, types.User]
}

func NewBootstrap(store protocols.Store[string, types.User]) *Bootstrap {
	return &Bootstrap{store: store}
}

func (uc *Bootstrap) Execute(ctx context.Context, name, token string) (types.User, error) {
	if token == "" {
		return types.User{}, auth.ErrUnauthorized
	}
	existing, err := uc.store.List(ctx, types.ListQuery{Page: types.Page{Limit: 1}})
	if err != nil {
		return types.User{}, err
	}
	if len(existing) > 0 {
		return existing[0], nil
	}
	user := types.User{
		ID:         uuid.NewString(),
		Name:       name,
		APIKeyHash: auth.HashToken(token),
		CreatedAt:  time.Now().UTC(),
	}
	created, err := uc.store.Create(ctx, []types.User{user})
	if err != nil {
		return types.User{}, err
	}
	return created[0], nil
}
