package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type DeleteSkill struct {
	store protocols.Store[string, types.Skill]
}

func NewDeleteSkill(store protocols.Store[string, types.Skill]) *DeleteSkill {
	return &DeleteSkill{store: store}
}

func (uc *DeleteSkill) Execute(ctx context.Context, id string) error {
	return uc.store.Delete(ctx, []string{id})
}
