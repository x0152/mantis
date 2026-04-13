package plans

import (
	"testing"
	"time"

	"mantis/core/types"
)

// --- renderPrompt ---

func TestRenderPrompt_NoInput(t *testing.T) {
	raw := "Hello {{.name}}"
	got := renderPrompt(raw, nil)
	if got != raw {
		t.Fatalf("expected unchanged prompt, got %q", got)
	}
}

func TestRenderPrompt_EmptyInput(t *testing.T) {
	raw := "Hello {{.name}}"
	got := renderPrompt(raw, map[string]any{})
	if got != raw {
		t.Fatalf("expected unchanged prompt, got %q", got)
	}
}

func TestRenderPrompt_NoTemplateMarkers(t *testing.T) {
	raw := "Just a plain prompt"
	got := renderPrompt(raw, map[string]any{"name": "world"})
	if got != raw {
		t.Fatalf("expected unchanged prompt, got %q", got)
	}
}

func TestRenderPrompt_SingleParam(t *testing.T) {
	got := renderPrompt("Hello {{.name}}!", map[string]any{"name": "Alice"})
	if got != "Hello Alice!" {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestRenderPrompt_MultipleParams(t *testing.T) {
	raw := "Search for {{.topic}} in {{.source}}"
	got := renderPrompt(raw, map[string]any{"topic": "AI", "source": "arxiv"})
	if got != "Search for AI in arxiv" {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestRenderPrompt_MissingKeyReturnsZeroValue(t *testing.T) {
	got := renderPrompt("Hello {{.name}}!", map[string]any{"other": "val"})
	// missingkey=zero renders <no value> for missing keys in map[string]any
	if got != "Hello <no value>!" {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestRenderPrompt_BadTemplate(t *testing.T) {
	raw := "Hello {{.name"
	got := renderPrompt(raw, map[string]any{"name": "x"})
	if got != raw {
		t.Fatalf("expected raw prompt on bad template, got %q", got)
	}
}

// --- parseDecision ---

func TestParseDecision_Yes(t *testing.T) {
	cases := []string{
		"yes", "Yes", "YES", "Yes, everything is fine", "yes.", "yes!",
	}
	for _, c := range cases {
		if got := parseDecision(c); got != "yes" {
			t.Errorf("parseDecision(%q) = %q, want yes", c, got)
		}
	}
}

func TestParseDecision_No(t *testing.T) {
	cases := []string{
		"no", "No", "NO", "No, there is a problem", "no.", "no!",
	}
	for _, c := range cases {
		if got := parseDecision(c); got != "no" {
			t.Errorf("parseDecision(%q) = %q, want no", c, got)
		}
	}
}

func TestParseDecision_EmptyDefaultsYes(t *testing.T) {
	if got := parseDecision(""); got != "yes" {
		t.Fatalf("expected yes for empty, got %q", got)
	}
}

func TestParseDecision_AmbiguousDefaultsYes(t *testing.T) {
	if got := parseDecision("maybe"); got != "yes" {
		t.Fatalf("expected yes for ambiguous, got %q", got)
	}
}

// --- validateGraph ---

func TestValidateGraph_Valid(t *testing.T) {
	g := types.PlanGraph{
		Nodes: []types.PlanNode{
			{ID: "n1", Type: types.PlanNodeAction},
			{ID: "n2", Type: types.PlanNodeDecision},
		},
		Edges: []types.PlanEdge{
			{ID: "e1", Source: "n1", Target: "n2"},
			{ID: "e2", Source: "n2", Target: "n1", Label: "yes"},
			{ID: "e3", Source: "n2", Target: "n1", Label: "no"},
		},
	}
	if err := validateGraph(g); err != nil {
		t.Fatal(err)
	}
}

func TestValidateGraph_ActionTooManyEdges(t *testing.T) {
	g := types.PlanGraph{
		Nodes: []types.PlanNode{
			{ID: "n1", Type: types.PlanNodeAction},
			{ID: "n2", Type: types.PlanNodeAction},
			{ID: "n3", Type: types.PlanNodeAction},
		},
		Edges: []types.PlanEdge{
			{ID: "e1", Source: "n1", Target: "n2"},
			{ID: "e2", Source: "n1", Target: "n3"},
		},
	}
	if err := validateGraph(g); err == nil {
		t.Fatal("expected error for action with 2 edges")
	}
}

func TestValidateGraph_DecisionTooManyEdges(t *testing.T) {
	g := types.PlanGraph{
		Nodes: []types.PlanNode{
			{ID: "n1", Type: types.PlanNodeDecision},
			{ID: "n2", Type: types.PlanNodeAction},
			{ID: "n3", Type: types.PlanNodeAction},
			{ID: "n4", Type: types.PlanNodeAction},
		},
		Edges: []types.PlanEdge{
			{ID: "e1", Source: "n1", Target: "n2", Label: "yes"},
			{ID: "e2", Source: "n1", Target: "n3", Label: "no"},
			{ID: "e3", Source: "n1", Target: "n4", Label: "maybe"},
		},
	}
	if err := validateGraph(g); err == nil {
		t.Fatal("expected error for decision with 3 edges")
	}
}

// --- findStartNodes ---

func TestFindStartNodes_Simple(t *testing.T) {
	g := types.PlanGraph{
		Nodes: []types.PlanNode{
			{ID: "n1"}, {ID: "n2"}, {ID: "n3"},
		},
		Edges: []types.PlanEdge{
			{Source: "n1", Target: "n2"},
			{Source: "n2", Target: "n3"},
		},
	}
	starts := findStartNodes(g)
	if len(starts) != 1 || starts[0] != "n1" {
		t.Fatalf("expected [n1], got %v", starts)
	}
}

func TestFindStartNodes_MultipleRoots(t *testing.T) {
	g := types.PlanGraph{
		Nodes: []types.PlanNode{
			{ID: "a"}, {ID: "b"}, {ID: "c"},
		},
		Edges: []types.PlanEdge{
			{Source: "a", Target: "c"},
		},
	}
	starts := findStartNodes(g)
	if len(starts) != 2 {
		t.Fatalf("expected 2 start nodes, got %v", starts)
	}
}

func TestFindStartNodes_NoEdges(t *testing.T) {
	g := types.PlanGraph{
		Nodes: []types.PlanNode{{ID: "n1"}, {ID: "n2"}},
	}
	starts := findStartNodes(g)
	if len(starts) != 2 {
		t.Fatalf("expected all nodes as starts, got %v", starts)
	}
}

// --- findNextNode ---

func TestFindNextNode_Found(t *testing.T) {
	g := types.PlanGraph{
		Edges: []types.PlanEdge{
			{Source: "n1", Target: "n2"},
		},
	}
	if got := findNextNode(g, "n1"); got != "n2" {
		t.Fatalf("expected n2, got %q", got)
	}
}

func TestFindNextNode_NotFound(t *testing.T) {
	g := types.PlanGraph{
		Edges: []types.PlanEdge{
			{Source: "n1", Target: "n2"},
		},
	}
	if got := findNextNode(g, "n3"); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

// --- findEdgeTarget ---

func TestFindEdgeTarget_ExactLabel(t *testing.T) {
	g := types.PlanGraph{
		Edges: []types.PlanEdge{
			{Source: "n1", Target: "n2", Label: "yes"},
			{Source: "n1", Target: "n3", Label: "no"},
		},
	}
	if got := findEdgeTarget(g, "n1", "yes"); got != "n2" {
		t.Fatalf("expected n2 for 'yes', got %q", got)
	}
	if got := findEdgeTarget(g, "n1", "no"); got != "n3" {
		t.Fatalf("expected n3 for 'no', got %q", got)
	}
}

func TestFindEdgeTarget_FallbackFirstEdge(t *testing.T) {
	g := types.PlanGraph{
		Edges: []types.PlanEdge{
			{Source: "n1", Target: "n2", Label: "yes"},
		},
	}
	if got := findEdgeTarget(g, "n1", "no"); got != "n2" {
		t.Fatalf("expected fallback to n2, got %q", got)
	}
}

func TestFindEdgeTarget_CaseInsensitive(t *testing.T) {
	g := types.PlanGraph{
		Edges: []types.PlanEdge{
			{Source: "n1", Target: "n2", Label: "Yes"},
		},
	}
	if got := findEdgeTarget(g, "n1", "YES"); got != "n2" {
		t.Fatalf("expected n2, got %q", got)
	}
}

// --- initSteps ---

func TestInitSteps(t *testing.T) {
	g := types.PlanGraph{
		Nodes: []types.PlanNode{
			{ID: "n1"}, {ID: "n2"}, {ID: "n3"},
		},
	}
	steps := initSteps(g)
	if len(steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(steps))
	}
	for _, s := range steps {
		if s.Status != "pending" {
			t.Fatalf("expected pending, got %q", s.Status)
		}
	}
}

// --- skipPending ---

func TestSkipPending(t *testing.T) {
	run := &types.PlanRun{
		Steps: []types.PlanStepRun{
			{NodeID: "n1", Status: "completed"},
			{NodeID: "n2", Status: "pending"},
			{NodeID: "n3", Status: "running"},
			{NodeID: "n4", Status: "pending"},
		},
	}
	skipPending(run)
	if run.Steps[0].Status != "completed" {
		t.Fatal("completed step should stay completed")
	}
	if run.Steps[1].Status != "skipped" {
		t.Fatal("pending step should become skipped")
	}
	if run.Steps[2].Status != "running" {
		t.Fatal("running step should stay running")
	}
	if run.Steps[3].Status != "skipped" {
		t.Fatal("pending step should become skipped")
	}
}

// --- markStaleSteps ---

func TestMarkStaleSteps(t *testing.T) {
	now := time.Now()
	run := &types.PlanRun{
		Steps: []types.PlanStepRun{
			{NodeID: "n1", Status: "completed"},
			{NodeID: "n2", Status: "running"},
			{NodeID: "n3", Status: "pending"},
		},
	}
	markStaleSteps(run, now, "shutdown")
	if run.Steps[0].Status != "completed" {
		t.Fatal("completed should remain")
	}
	if run.Steps[1].Status != "failed" {
		t.Fatal("running should become failed")
	}
	if run.Steps[1].Result != "shutdown" {
		t.Fatal("running step should have reason")
	}
	if run.Steps[2].Status != "skipped" {
		t.Fatal("pending should become skipped")
	}
}

// --- decisionPrompt ---

func TestDecisionPrompt_WithClearContext(t *testing.T) {
	node := types.PlanNode{
		Prompt:       "Is {{.service}} healthy?",
		ClearContext: true,
	}
	got := decisionPrompt(node, map[string]any{"service": "nginx"})
	if got == "" {
		t.Fatal("expected non-empty prompt")
	}
	if !contains(got, "Is nginx healthy?") {
		t.Fatalf("expected rendered prompt, got %q", got)
	}
	if !contains(got, "EXACTLY 'yes' or 'no'") {
		t.Fatalf("expected yes/no instruction, got %q", got)
	}
}

func TestDecisionPrompt_WithoutClearContext(t *testing.T) {
	node := types.PlanNode{Prompt: "Continue?"}
	got := decisionPrompt(node, nil)
	if !contains(got, "Based on everything above") {
		t.Fatalf("expected context-aware prefix, got %q", got)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
