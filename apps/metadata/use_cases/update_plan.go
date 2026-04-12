package usecases

import (
	"context"
	"strings"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type UpdatePlan struct {
	store protocols.Store[string, types.Plan]
}

func NewUpdatePlan(store protocols.Store[string, types.Plan]) *UpdatePlan {
	return &UpdatePlan{store: store}
}

func (uc *UpdatePlan) Execute(ctx context.Context, p types.Plan) (types.Plan, error) {
	existing, err := uc.store.Get(ctx, []string{p.ID})
	if err != nil {
		return types.Plan{}, err
	}
	if _, ok := existing[p.ID]; !ok {
		return types.Plan{}, base.ErrNotFound
	}
	p = normalizePlan(p)
	if strings.TrimSpace(p.Name) == "" {
		return types.Plan{}, base.ErrValidation
	}
	result, err := uc.store.Update(ctx, []types.Plan{p})
	if err != nil {
		return types.Plan{}, err
	}
	return result[0], nil
}
