package usecases

import (
	"context"

	"mantis/core/protocols"
	"mantis/core/types"
)

type ListPlanRuns struct {
	store protocols.Store[string, types.PlanRun]
}

func NewListPlanRuns(store protocols.Store[string, types.PlanRun]) *ListPlanRuns {
	return &ListPlanRuns{store: store}
}

func (uc *ListPlanRuns) Execute(ctx context.Context, planID string) ([]types.PlanRun, error) {
	query := types.ListQuery{
		Sort: []types.Sort{{Field: "started_at", Dir: types.SortDirDesc}},
	}
	if planID != "" {
		query.Filter = map[string]string{"plan_id": planID}
	}
	items, err := uc.store.List(ctx, query)
	if items == nil {
		items = []types.PlanRun{}
	}
	return items, err
}
