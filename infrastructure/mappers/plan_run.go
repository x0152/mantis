package mappers

import (
	"encoding/json"

	"mantis/core/types"
	"mantis/infrastructure/models"
)

func PlanRunToRow(r types.PlanRun) models.PlanRunRow {
	steps, _ := json.Marshal(r.Steps)
	input, _ := json.Marshal(r.Input)
	if len(input) == 0 || string(input) == "null" {
		input = json.RawMessage(`{}`)
	}
	return models.PlanRunRow{
		ID:         r.ID,
		PlanID:     r.PlanID,
		Status:     r.Status,
		Trigger:    r.Trigger,
		Input:      input,
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
	var input map[string]any
	_ = json.Unmarshal(r.Input, &input)
	if input == nil {
		input = map[string]any{}
	}
	return types.PlanRun{
		ID:         r.ID,
		PlanID:     r.PlanID,
		Status:     r.Status,
		Trigger:    r.Trigger,
		Input:      input,
		Steps:      steps,
		StartedAt:  r.StartedAt,
		FinishedAt: r.FinishedAt,
	}
}
