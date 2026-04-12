package usecases

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"mantis/core/base"
	"mantis/core/protocols"
	"mantis/core/types"
)

type CreatePlan struct {
	store protocols.Store[string, types.Plan]
}

func NewCreatePlan(store protocols.Store[string, types.Plan]) *CreatePlan {
	return &CreatePlan{store: store}
}

func (uc *CreatePlan) Execute(ctx context.Context, p types.Plan) (types.Plan, error) {
	p.ID = uuid.New().String()
	p = normalizePlan(p)
	if strings.TrimSpace(p.Name) == "" {
		return types.Plan{}, base.ErrValidation
	}
	result, err := uc.store.Create(ctx, []types.Plan{p})
	if err != nil {
		return types.Plan{}, err
	}
	return result[0], nil
}

func normalizePlan(p types.Plan) types.Plan {
	p.Name = strings.TrimSpace(p.Name)
	p.Description = strings.TrimSpace(p.Description)
	p.Schedule = strings.TrimSpace(p.Schedule)
	if p.Graph.Nodes == nil {
		p.Graph.Nodes = []types.PlanNode{}
	}
	if p.Graph.Edges == nil {
		p.Graph.Edges = []types.PlanEdge{}
	}
	return p
}
