package types

import "encoding/json"

type PlanNodeType string

const (
	PlanNodeAction   PlanNodeType = "action"
	PlanNodeDecision PlanNodeType = "decision"
)

type PlanNode struct {
	ID           string          `json:"id"`
	Type         PlanNodeType    `json:"type"`
	Label        string          `json:"label"`
	Prompt       string          `json:"prompt"`
	Position     json.RawMessage `json:"position"`
	ClearContext bool            `json:"clearContext,omitempty"`
	MaxRetries   int             `json:"maxRetries,omitempty"`
}

type PlanEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"`
}

type PlanGraph struct {
	Nodes []PlanNode `json:"nodes"`
	Edges []PlanEdge `json:"edges"`
}

type Plan struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Schedule    string    `json:"schedule"`
	Enabled     bool      `json:"enabled"`
	Graph       PlanGraph `json:"graph"`
}
