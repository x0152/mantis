package usecases

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type CreateSkill struct {
	store protocols.Store[string, types.Skill]
}

func NewCreateSkill(store protocols.Store[string, types.Skill]) *CreateSkill {
	return &CreateSkill{store: store}
}

func (uc *CreateSkill) Execute(ctx context.Context, s types.Skill) (types.Skill, error) {
	s.ID = uuid.New().String()
	s = normalizeSkill(s)
	if strings.TrimSpace(s.ConnectionID) == "" || strings.TrimSpace(s.Name) == "" || strings.TrimSpace(s.Script) == "" {
		return types.Skill{}, base.ErrValidation
	}
	result, err := uc.store.Create(ctx, []types.Skill{s})
	if err != nil {
		return types.Skill{}, err
	}
	return result[0], nil
}

func normalizeSkill(s types.Skill) types.Skill {
	s.ConnectionID = strings.TrimSpace(s.ConnectionID)
	s.Name = strings.TrimSpace(s.Name)
	s.Description = strings.TrimSpace(s.Description)
	if len(s.Parameters) == 0 || strings.TrimSpace(string(s.Parameters)) == "" || strings.TrimSpace(string(s.Parameters)) == "null" {
		s.Parameters = json.RawMessage(`{}`)
	}
	return s
}
