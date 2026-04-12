package usecases

import (
	"context"
	"strings"

	"mantis/core/protocols"
	"mantis/core/types"
)

type ListSkills struct {
	store protocols.Store[string, types.Skill]
}

func NewListSkills(store protocols.Store[string, types.Skill]) *ListSkills {
	return &ListSkills{store: store}
}

func (uc *ListSkills) Execute(ctx context.Context, connectionID string) ([]types.Skill, error) {
	query := types.ListQuery{
		Sort: []types.Sort{{Field: "name", Dir: types.SortDirAsc}},
	}
	if strings.TrimSpace(connectionID) != "" {
		query.Filter = map[string]string{"connection_id": strings.TrimSpace(connectionID)}
	}
	items, err := uc.store.List(ctx, query)
	if items == nil {
		items = []types.Skill{}
	}
	return items, err
}
