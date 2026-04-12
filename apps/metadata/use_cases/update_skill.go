package usecases

import (
	"context"
	"strings"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type UpdateSkill struct {
	store protocols.Store[string, types.Skill]
}

func NewUpdateSkill(store protocols.Store[string, types.Skill]) *UpdateSkill {
	return &UpdateSkill{store: store}
}

func (uc *UpdateSkill) Execute(ctx context.Context, s types.Skill) (types.Skill, error) {
	existing, err := uc.store.Get(ctx, []string{s.ID})
	if err != nil {
		return types.Skill{}, err
	}
	if _, ok := existing[s.ID]; !ok {
		return types.Skill{}, base.ErrNotFound
	}
	s = normalizeSkill(s)
	if strings.TrimSpace(s.ConnectionID) == "" || strings.TrimSpace(s.Name) == "" || strings.TrimSpace(s.Script) == "" {
		return types.Skill{}, base.ErrValidation
	}
	result, err := uc.store.Update(ctx, []types.Skill{s})
	if err != nil {
		return types.Skill{}, err
	}
	return result[0], nil
}
