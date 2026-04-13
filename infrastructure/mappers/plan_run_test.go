package mappers

import (
	"encoding/json"
	"testing"
	"time"

	"mantis/core/types"
	"mantis/infrastructure/models"
)

func TestPlanRunToRow_Basic(t *testing.T) {
	now := time.Now().UTC()
	run := types.PlanRun{
		ID:        "r1",
		PlanID:    "p1",
		Status:    "completed",
		Trigger:   "manual",
		Input:     map[string]any{"topic": "AI", "depth": float64(3)},
		Steps:     []types.PlanStepRun{{NodeID: "n1", Status: "completed"}},
		StartedAt: now,
	}
	row := PlanRunToRow(run)
	if row.ID != "r1" {
		t.Fatalf("ID: %s", row.ID)
	}
	if row.PlanID != "p1" {
		t.Fatalf("PlanID: %s", row.PlanID)
	}

	var input map[string]any
	if err := json.Unmarshal(row.Input, &input); err != nil {
		t.Fatalf("Input unmarshal: %v", err)
	}
	if input["topic"] != "AI" {
		t.Fatalf("Input topic: %v", input["topic"])
	}

	var steps []types.PlanStepRun
	if err := json.Unmarshal(row.Steps, &steps); err != nil {
		t.Fatalf("Steps unmarshal: %v", err)
	}
	if len(steps) != 1 || steps[0].NodeID != "n1" {
		t.Fatalf("Steps mismatch")
	}
}

func TestPlanRunToRow_NilInput(t *testing.T) {
	run := types.PlanRun{ID: "r1"}
	row := PlanRunToRow(run)
	if string(row.Input) != `{}` {
		t.Fatalf("expected '{}' for nil input, got %s", string(row.Input))
	}
}

func TestPlanRunFromRow_Basic(t *testing.T) {
	now := time.Now().UTC()
	steps, _ := json.Marshal([]types.PlanStepRun{
		{NodeID: "n1", Status: "completed"},
		{NodeID: "n2", Status: "failed", Result: "error"},
	})
	input, _ := json.Marshal(map[string]any{"service": "nginx"})
	row := models.PlanRunRow{
		ID:        "r1",
		PlanID:    "p1",
		Status:    "failed",
		Trigger:   "schedule",
		Input:     input,
		Steps:     steps,
		StartedAt: now,
	}
	run := PlanRunFromRow(row)
	if run.ID != "r1" || run.PlanID != "p1" {
		t.Fatal("basic field mismatch")
	}
	if run.Status != "failed" {
		t.Fatalf("Status: %s", run.Status)
	}
	if run.Input["service"] != "nginx" {
		t.Fatalf("Input service: %v", run.Input["service"])
	}
	if len(run.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(run.Steps))
	}
	if run.Steps[1].Result != "error" {
		t.Fatalf("Step result: %s", run.Steps[1].Result)
	}
}

func TestPlanRunFromRow_NilFields(t *testing.T) {
	row := models.PlanRunRow{ID: "r1"}
	run := PlanRunFromRow(row)
	if run.Steps == nil {
		t.Fatal("Steps should be empty slice, not nil")
	}
	if run.Input == nil {
		t.Fatal("Input should be empty map, not nil")
	}
}

func TestPlanRunRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	finished := now.Add(5 * time.Minute)
	original := types.PlanRun{
		ID:      "run-456",
		PlanID:  "plan-123",
		Status:  "completed",
		Trigger: "api",
		Input:   map[string]any{"topic": "Rust", "limit": float64(10)},
		Steps: []types.PlanStepRun{
			{NodeID: "n1", Status: "completed", Result: "done"},
			{NodeID: "n2", Status: "skipped"},
		},
		StartedAt:  now,
		FinishedAt: &finished,
	}

	row := PlanRunToRow(original)
	restored := PlanRunFromRow(row)

	if restored.ID != original.ID {
		t.Fatal("ID")
	}
	if restored.PlanID != original.PlanID {
		t.Fatal("PlanID")
	}
	if restored.Status != original.Status {
		t.Fatal("Status")
	}
	if restored.Trigger != original.Trigger {
		t.Fatal("Trigger")
	}
	if restored.Input["topic"] != original.Input["topic"] {
		t.Fatalf("Input topic: %v vs %v", restored.Input["topic"], original.Input["topic"])
	}
	if restored.Input["limit"] != original.Input["limit"] {
		t.Fatalf("Input limit: %v vs %v", restored.Input["limit"], original.Input["limit"])
	}
	if len(restored.Steps) != len(original.Steps) {
		t.Fatal("Steps count")
	}
	if restored.Steps[0].Result != original.Steps[0].Result {
		t.Fatal("Step result")
	}
	if restored.Steps[1].Status != "skipped" {
		t.Fatal("Step 2 status")
	}
}
