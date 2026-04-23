package usecases

import (
	"context"

	"mantis/core/auth"
	"mantis/core/protocols"
	"mantis/core/types"
)

type Me struct {
	store protocols.Store[string, types.User]
}

func NewMe(store protocols.Store[string, types.User]) *Me {
	return &Me{store: store}
}

func (uc *Me) Execute(ctx context.Context) (types.User, error) {
	id, ok := auth.FromContext(ctx)
	if !ok {
		return types.User{}, auth.ErrUnauthorized
	}
	users, err := uc.store.Get(ctx, []string{id.UserID})
	if err != nil {
		return types.User{}, err
	}
	user, ok := users[id.UserID]
	if !ok {
		return types.User{}, auth.ErrUnauthorized
	}
	return user, nil
}
