package agents

import (
	"context"
	"encoding/json"
	"testing"

	"mantis/core/types"
)

type planStoreMock struct {
	plans   map[string]types.Plan
	updated []types.Plan
}

func (s *planStoreMock) Create(_ context.Context, items []types.Plan) ([]types.Plan, error) {
	for _, p := range items {
		s.plans[p.ID] = p
	}
	return items, nil
}

func (s *planStoreMock) Get(_ context.Context, ids []string) (map[string]types.Plan, error) {
	out := make(map[string]types.Plan)
	for _, id := range ids {
		if p, ok := s.plans[id]; ok {
			out[id] = p
		}
	}
	return out, nil
}

func (s *planStoreMock) List(_ context.Context, _ types.ListQuery) ([]types.Plan, error) {
	var out []types.Plan
	for _, p := range s.plans {
		out = append(out, p)
	}
	return out, nil
}

func (s *planStoreMock) Update(_ context.Context, items []types.Plan) ([]types.Plan, error) {
	s.updated = items
	for _, p := range items {
		s.plans[p.ID] = p
	}
	return items, nil
}

func (s *planStoreMock) Delete(_ context.Context, ids []string) error {
	for _, id := range ids {
		delete(s.plans, id)
	}
	return nil
}

func newAgentWithPlanStore(store *planStoreMock) *MantisAgent {
	return &MantisAgent{planStore: store}
}

func TestPlanUpdateTool_ChangeSchedule(t *testing.T) {
	store := &planStoreMock{plans: map[string]types.Plan{
		"p1": {ID: "p1", Name: "My Plan", Schedule: "0 9 * * *", Enabled: true},
	}}
	agent := newAgentWithPlanStore(store)
	tool := agent.planUpdateTool()

	result, err := tool.Execute(context.Background(), `{"id":"p1","schedule":"0 */6 * * *"}`)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	_ = json.Unmarshal([]byte(result), &out)
	if out["schedule"] != "0 */6 * * *" {
		t.Fatalf("expected new schedule, got %v", out["schedule"])
	}
	if store.plans["p1"].Schedule != "0 */6 * * *" {
		t.Fatal("store not updated")
	}
	if store.plans["p1"].Name != "My Plan" {
		t.Fatal("name should be unchanged")
	}
}

func TestPlanUpdateTool_RemoveSchedule(t *testing.T) {
	store := &planStoreMock{plans: map[string]types.Plan{
		"p1": {ID: "p1", Schedule: "0 9 * * *"},
	}}
	agent := newAgentWithPlanStore(store)
	tool := agent.planUpdateTool()

	_, err := tool.Execute(context.Background(), `{"id":"p1","schedule":""}`)
	if err != nil {
		t.Fatal(err)
	}
	if store.plans["p1"].Schedule != "" {
		t.Fatalf("expected empty schedule, got %q", store.plans["p1"].Schedule)
	}
}

func TestPlanUpdateTool_Disable(t *testing.T) {
	store := &planStoreMock{plans: map[string]types.Plan{
		"p1": {ID: "p1", Enabled: true},
	}}
	agent := newAgentWithPlanStore(store)
	tool := agent.planUpdateTool()

	_, err := tool.Execute(context.Background(), `{"id":"p1","enabled":false}`)
	if err != nil {
		t.Fatal(err)
	}
	if store.plans["p1"].Enabled {
		t.Fatal("expected disabled")
	}
}

func TestPlanUpdateTool_Enable(t *testing.T) {
	store := &planStoreMock{plans: map[string]types.Plan{
		"p1": {ID: "p1", Enabled: false},
	}}
	agent := newAgentWithPlanStore(store)
	tool := agent.planUpdateTool()

	_, err := tool.Execute(context.Background(), `{"id":"p1","enabled":true}`)
	if err != nil {
		t.Fatal(err)
	}
	if !store.plans["p1"].Enabled {
		t.Fatal("expected enabled")
	}
}

func TestPlanUpdateTool_Rename(t *testing.T) {
	store := &planStoreMock{plans: map[string]types.Plan{
		"p1": {ID: "p1", Name: "Old Name"},
	}}
	agent := newAgentWithPlanStore(store)
	tool := agent.planUpdateTool()

	_, err := tool.Execute(context.Background(), `{"id":"p1","name":"New Name"}`)
	if err != nil {
		t.Fatal(err)
	}
	if store.plans["p1"].Name != "New Name" {
		t.Fatalf("expected 'New Name', got %q", store.plans["p1"].Name)
	}
}

func TestPlanUpdateTool_EmptyNameIgnored(t *testing.T) {
	store := &planStoreMock{plans: map[string]types.Plan{
		"p1": {ID: "p1", Name: "Keep Me"},
	}}
	agent := newAgentWithPlanStore(store)
	tool := agent.planUpdateTool()

	_, err := tool.Execute(context.Background(), `{"id":"p1","name":"  "}`)
	if err != nil {
		t.Fatal(err)
	}
	if store.plans["p1"].Name != "Keep Me" {
		t.Fatal("empty name should be ignored")
	}
}

func TestPlanUpdateTool_UpdateDescription(t *testing.T) {
	store := &planStoreMock{plans: map[string]types.Plan{
		"p1": {ID: "p1", Description: "old"},
	}}
	agent := newAgentWithPlanStore(store)
	tool := agent.planUpdateTool()

	_, err := tool.Execute(context.Background(), `{"id":"p1","description":"new description"}`)
	if err != nil {
		t.Fatal(err)
	}
	if store.plans["p1"].Description != "new description" {
		t.Fatal("description not updated")
	}
}

func TestPlanUpdateTool_MultipleFields(t *testing.T) {
	store := &planStoreMock{plans: map[string]types.Plan{
		"p1": {ID: "p1", Name: "Old", Schedule: "0 9 * * *", Enabled: false},
	}}
	agent := newAgentWithPlanStore(store)
	tool := agent.planUpdateTool()

	_, err := tool.Execute(context.Background(), `{"id":"p1","name":"New","schedule":"0 12 * * *","enabled":true}`)
	if err != nil {
		t.Fatal(err)
	}
	p := store.plans["p1"]
	if p.Name != "New" || p.Schedule != "0 12 * * *" || !p.Enabled {
		t.Fatalf("multiple fields not updated: %+v", p)
	}
}

func TestPlanUpdateTool_PlanNotFound(t *testing.T) {
	store := &planStoreMock{plans: map[string]types.Plan{}}
	agent := newAgentWithPlanStore(store)
	tool := agent.planUpdateTool()

	_, err := tool.Execute(context.Background(), `{"id":"missing","enabled":true}`)
	if err == nil {
		t.Fatal("expected error for missing plan")
	}
}

func TestPlanUpdateTool_EmptyID(t *testing.T) {
	store := &planStoreMock{plans: map[string]types.Plan{}}
	agent := newAgentWithPlanStore(store)
	tool := agent.planUpdateTool()

	_, err := tool.Execute(context.Background(), `{"id":""}`)
	if err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestPlanUpdateTool_NilStore(t *testing.T) {
	agent := &MantisAgent{}
	tool := agent.planUpdateTool()

	_, err := tool.Execute(context.Background(), `{"id":"p1"}`)
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestPlanUpdateTool_OnlyIDNoChanges(t *testing.T) {
	store := &planStoreMock{plans: map[string]types.Plan{
		"p1": {ID: "p1", Name: "Same", Schedule: "0 9 * * *", Enabled: true},
	}}
	agent := newAgentWithPlanStore(store)
	tool := agent.planUpdateTool()

	_, err := tool.Execute(context.Background(), `{"id":"p1"}`)
	if err != nil {
		t.Fatal(err)
	}
	p := store.plans["p1"]
	if p.Name != "Same" || p.Schedule != "0 9 * * *" || !p.Enabled {
		t.Fatal("nothing should change")
	}
}

func TestPlanUpdateTool_Labels(t *testing.T) {
	agent := &MantisAgent{}
	tool := agent.planUpdateTool()

	if label := tool.Label(`{"schedule":"0 9 * * *"}`); label != "Update plan schedule" {
		t.Fatalf("expected schedule label, got %q", label)
	}
	if label := tool.Label(`{"enabled":false}`); label != "Disable plan" {
		t.Fatalf("expected disable label, got %q", label)
	}
	if label := tool.Label(`{"enabled":true}`); label != "Enable plan" {
		t.Fatalf("expected enable label, got %q", label)
	}
	if label := tool.Label(`{"name":"x"}`); label != "Update plan" {
		t.Fatalf("expected generic label, got %q", label)
	}
}
