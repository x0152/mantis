package agents

import (
	"testing"

	"mantis/core/types"
)

func TestStepsToGraph_SingleAction(t *testing.T) {
	steps := []planStep{
		{Type: "action", Prompt: "Do something"},
	}
	g, err := stepsToGraph(steps)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(g.Nodes))
	}
	if g.Nodes[0].Type != types.PlanNodeAction {
		t.Fatalf("expected action, got %s", g.Nodes[0].Type)
	}
	if g.Nodes[0].Prompt != "Do something" {
		t.Fatalf("unexpected prompt: %s", g.Nodes[0].Prompt)
	}
	if len(g.Edges) != 0 {
		t.Fatalf("expected 0 edges for single node, got %d", len(g.Edges))
	}
}

func TestStepsToGraph_LinearChain(t *testing.T) {
	steps := []planStep{
		{Type: "action", Prompt: "Step 1"},
		{Type: "action", Prompt: "Step 2"},
		{Type: "action", Prompt: "Step 3"},
	}
	g, err := stepsToGraph(steps)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(g.Edges))
	}
	if g.Edges[0].Source != "n1" || g.Edges[0].Target != "n2" {
		t.Fatalf("edge 0: expected n1->n2, got %s->%s", g.Edges[0].Source, g.Edges[0].Target)
	}
	if g.Edges[1].Source != "n2" || g.Edges[1].Target != "n3" {
		t.Fatalf("edge 1: expected n2->n3, got %s->%s", g.Edges[1].Source, g.Edges[1].Target)
	}
}

func TestStepsToGraph_DecisionBranching(t *testing.T) {
	steps := []planStep{
		{Type: "action", Prompt: "Check health"},
		{Type: "decision", Prompt: "Any issues?", Yes: "next", No: "ok"},
		{Type: "action", Prompt: "Send alert"},
		{Type: "action", Prompt: "All good", ID: "ok"},
	}
	g, err := stepsToGraph(steps)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(g.Nodes))
	}
	if g.Nodes[1].Type != types.PlanNodeDecision {
		t.Fatalf("node 1 should be decision, got %s", g.Nodes[1].Type)
	}

	edgeMap := make(map[string]string)
	for _, e := range g.Edges {
		edgeMap[e.Source+":"+e.Label] = e.Target
	}
	// n1 -> n2 (action auto-chain)
	if edgeMap["n1:"] != "n2" {
		t.Fatalf("expected n1->n2, got %v", edgeMap)
	}
	// n2 yes -> n3
	if edgeMap["n2:yes"] != "n3" {
		t.Fatalf("expected n2 yes->n3, got %v", edgeMap)
	}
	// n2 no -> n4 (the "ok" step)
	if edgeMap["n2:no"] != "n4" {
		t.Fatalf("expected n2 no->n4, got %v", edgeMap)
	}
}

func TestStepsToGraph_DecisionEndTarget(t *testing.T) {
	steps := []planStep{
		{Type: "decision", Prompt: "Continue?", Yes: "next", No: "end"},
		{Type: "action", Prompt: "Do work"},
	}
	g, err := stepsToGraph(steps)
	if err != nil {
		t.Fatal(err)
	}

	var yesEdge, noEdge bool
	for _, e := range g.Edges {
		if e.Source == "n1" && e.Label == "yes" && e.Target == "n2" {
			yesEdge = true
		}
		if e.Source == "n1" && e.Label == "no" {
			noEdge = true
		}
	}
	if !yesEdge {
		t.Fatal("expected yes edge from n1 to n2")
	}
	if noEdge {
		t.Fatal("expected no 'no' edge (target was 'end')")
	}
}

func TestStepsToGraph_AutoLabels(t *testing.T) {
	steps := []planStep{
		{Type: "action", Prompt: "Do"},
		{Type: "decision", Prompt: "Check?"},
	}
	g, err := stepsToGraph(steps)
	if err != nil {
		t.Fatal(err)
	}
	if g.Nodes[0].Label != "Step 1" {
		t.Fatalf("expected auto-label 'Step 1', got %q", g.Nodes[0].Label)
	}
	if g.Nodes[1].Label != "Decision 2" {
		t.Fatalf("expected auto-label 'Decision 2', got %q", g.Nodes[1].Label)
	}
}

func TestStepsToGraph_CustomLabels(t *testing.T) {
	steps := []planStep{
		{Type: "action", Prompt: "Deploy", Label: "Deploy to prod"},
	}
	g, err := stepsToGraph(steps)
	if err != nil {
		t.Fatal(err)
	}
	if g.Nodes[0].Label != "Deploy to prod" {
		t.Fatalf("expected custom label, got %q", g.Nodes[0].Label)
	}
}

func TestStepsToGraph_Empty(t *testing.T) {
	_, err := stepsToGraph(nil)
	if err == nil {
		t.Fatal("expected error for empty steps")
	}
}

func TestStepsToGraph_TooManySteps(t *testing.T) {
	steps := make([]planStep, maxAgentPlanSteps+1)
	for i := range steps {
		steps[i] = planStep{Type: "action", Prompt: "x"}
	}
	_, err := stepsToGraph(steps)
	if err == nil {
		t.Fatal("expected error for too many steps")
	}
}

func TestStepsToGraph_NodeIDsUnique(t *testing.T) {
	steps := []planStep{
		{Type: "action", Prompt: "A"},
		{Type: "action", Prompt: "B"},
		{Type: "action", Prompt: "C"},
	}
	g, err := stepsToGraph(steps)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, n := range g.Nodes {
		if seen[n.ID] {
			t.Fatalf("duplicate node ID: %s", n.ID)
		}
		seen[n.ID] = true
	}
}

func TestStepsToGraph_EdgeIDsUnique(t *testing.T) {
	steps := []planStep{
		{Type: "action", Prompt: "A"},
		{Type: "decision", Prompt: "B?", Yes: "next", No: "end"},
		{Type: "action", Prompt: "C"},
	}
	g, err := stepsToGraph(steps)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, e := range g.Edges {
		if seen[e.ID] {
			t.Fatalf("duplicate edge ID: %s", e.ID)
		}
		seen[e.ID] = true
	}
}
