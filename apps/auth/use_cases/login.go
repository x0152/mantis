package usecases

import (
	"context"

	"mantis/core/auth"
	"mantis/core/protocols"
	"mantis/core/types"
)

type Login struct {
	store protocols.Store[string, types.User]
}

func NewLogin(store protocols.Store[string, types.User]) *Login {
	return &Login{store: store}
}

func (uc *Login) Execute(ctx context.Context, token string) (types.User, error) {
	if token == "" {
		return types.User{}, auth.ErrUnauthorized
	}
	users, err := uc.store.List(ctx, types.ListQuery{
		Filter: map[string]string{"api_key_hash": auth.HashToken(token)},
		Page:   types.Page{Limit: 1},
	})
	if err != nil {
		return types.User{}, err
	}
	if len(users) == 0 {
		return types.User{}, auth.ErrUnauthorized
	}
	return users[0], nil
}
