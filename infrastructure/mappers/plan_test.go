package mappers

import (
	"encoding/json"
	"testing"

	"mantis/core/types"
	"mantis/infrastructure/models"
)

func TestPlanToRow_Basic(t *testing.T) {
	plan := types.Plan{
		ID:          "p1",
		Name:        "Test",
		Description: "desc",
		Schedule:    "0 * * * *",
		Enabled:     true,
		Parameters:  json.RawMessage(`{"properties":{"topic":{"type":"string"}}}`),
		Graph: types.PlanGraph{
			Nodes: []types.PlanNode{{ID: "n1", Type: types.PlanNodeAction, Label: "Do"}},
			Edges: []types.PlanEdge{},
		},
	}
	row := PlanToRow(plan)
	if row.ID != "p1" {
		t.Fatalf("ID mismatch: %s", row.ID)
	}
	if row.Name != "Test" {
		t.Fatalf("Name mismatch: %s", row.Name)
	}
	if row.Schedule != "0 * * * *" {
		t.Fatalf("Schedule mismatch: %s", row.Schedule)
	}
	if !row.Enabled {
		t.Fatal("Enabled should be true")
	}
	if string(row.Parameters) != `{"properties":{"topic":{"type":"string"}}}` {
		t.Fatalf("Parameters mismatch: %s", string(row.Parameters))
	}
	var graph types.PlanGraph
	if err := json.Unmarshal(row.Graph, &graph); err != nil {
		t.Fatalf("Graph unmarshal: %v", err)
	}
	if len(graph.Nodes) != 1 {
		t.Fatalf("expected 1 node in graph, got %d", len(graph.Nodes))
	}
}

func TestPlanToRow_EmptyParameters(t *testing.T) {
	plan := types.Plan{ID: "p1"}
	row := PlanToRow(plan)
	if string(row.Parameters) != `{}` {
		t.Fatalf("expected '{}' for empty parameters, got %s", string(row.Parameters))
	}
}

func TestPlanFromRow_Basic(t *testing.T) {
	graph, _ := json.Marshal(types.PlanGraph{
		Nodes: []types.PlanNode{{ID: "n1", Type: types.PlanNodeAction}},
		Edges: []types.PlanEdge{{ID: "e1", Source: "n1", Target: "n2"}},
	})
	row := models.PlanRow{
		ID:          "p1",
		Name:        "Test",
		Description: "desc",
		Schedule:    "*/5 * * * *",
		Enabled:     true,
		Parameters:  json.RawMessage(`{"properties":{"x":{"type":"number"}}}`),
		Graph:       graph,
	}
	plan := PlanFromRow(row)
	if plan.ID != "p1" || plan.Name != "Test" {
		t.Fatalf("basic field mismatch")
	}
	if len(plan.Graph.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(plan.Graph.Nodes))
	}
	if len(plan.Graph.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(plan.Graph.Edges))
	}
	if string(plan.Parameters) != `{"properties":{"x":{"type":"number"}}}` {
		t.Fatalf("Parameters mismatch: %s", string(plan.Parameters))
	}
}

func TestPlanFromRow_NilGraph(t *testing.T) {
	row := models.PlanRow{ID: "p1"}
	plan := PlanFromRow(row)
	if plan.Graph.Nodes == nil {
		t.Fatal("Nodes should be empty slice, not nil")
	}
	if plan.Graph.Edges == nil {
		t.Fatal("Edges should be empty slice, not nil")
	}
}

func TestPlanFromRow_EmptyParameters(t *testing.T) {
	row := models.PlanRow{ID: "p1"}
	plan := PlanFromRow(row)
	if string(plan.Parameters) != `{}` {
		t.Fatalf("expected '{}', got %s", string(plan.Parameters))
	}
}

func TestPlanRoundTrip(t *testing.T) {
	original := types.Plan{
		ID:          "plan-123",
		Name:        "My Plan",
		Description: "A test plan",
		Schedule:    "0 9 * * 1",
		Enabled:     true,
		Parameters:  json.RawMessage(`{"type":"object","properties":{"topic":{"type":"string"}}}`),
		Graph: types.PlanGraph{
			Nodes: []types.PlanNode{
				{ID: "n1", Type: types.PlanNodeAction, Label: "Research", Prompt: "Search {{.topic}}"},
				{ID: "n2", Type: types.PlanNodeDecision, Label: "Enough?", Prompt: "Got enough info?"},
			},
			Edges: []types.PlanEdge{
				{ID: "e1", Source: "n1", Target: "n2"},
				{ID: "e2", Source: "n2", Target: "n1", Label: "no"},
			},
		},
	}

	row := PlanToRow(original)
	restored := PlanFromRow(row)

	if restored.ID != original.ID {
		t.Fatal("ID mismatch")
	}
	if restored.Name != original.Name {
		t.Fatal("Name mismatch")
	}
	if restored.Description != original.Description {
		t.Fatal("Description mismatch")
	}
	if restored.Schedule != original.Schedule {
		t.Fatal("Schedule mismatch")
	}
	if restored.Enabled != original.Enabled {
		t.Fatal("Enabled mismatch")
	}
	if string(restored.Parameters) != string(original.Parameters) {
		t.Fatalf("Parameters mismatch: %s vs %s", string(restored.Parameters), string(original.Parameters))
	}
	if len(restored.Graph.Nodes) != len(original.Graph.Nodes) {
		t.Fatal("Nodes count mismatch")
	}
	if len(restored.Graph.Edges) != len(original.Graph.Edges) {
		t.Fatal("Edges count mismatch")
	}
	if restored.Graph.Nodes[0].Prompt != original.Graph.Nodes[0].Prompt {
		t.Fatalf("Prompt mismatch: %q vs %q", restored.Graph.Nodes[0].Prompt, original.Graph.Nodes[0].Prompt)
	}
}
