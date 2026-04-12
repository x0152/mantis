package mappers

import (
	"encoding/json"

	"mantis/core/types"
	"mantis/infrastructure/models"
)

func PlanRunToRow(r types.PlanRun) models.PlanRunRow {
	steps, _ := json.Marshal(r.Steps)
	return models.PlanRunRow{
		ID:         r.ID,
		PlanID:     r.PlanID,
		Status:     r.Status,
		Trigger:    r.Trigger,
		Steps:      steps,
		StartedAt:  r.StartedAt,
		FinishedAt: r.FinishedAt,
	}
}

func PlanRunFromRow(r models.PlanRunRow) types.PlanRun {
	var steps []types.PlanStepRun
	_ = json.Unmarshal(r.Steps, &steps)
	if steps == nil {
		steps = []types.PlanStepRun{}
	}
	return types.PlanRun{
		ID:         r.ID,
		PlanID:     r.PlanID,
		Status:     r.Status,
		Trigger:    r.Trigger,
		Steps:      steps,
		StartedAt:  r.StartedAt,
		FinishedAt: r.FinishedAt,
	}
}
