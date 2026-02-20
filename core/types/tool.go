package types

import "context"

type Tool struct {
	Name        string
	Description string
	Icon        string
	Label       func(args string) string
	Parameters  map[string]any
	Execute     func(ctx context.Context, args string) (string, error)
}

type Step struct {
	ID            string `json:"id"`
	Tool          string `json:"tool"`
	Label         string `json:"label"`
	Icon          string `json:"icon"`
	Args          string `json:"args"`
	Status        string `json:"status"`
	Result        string `json:"result,omitempty"`
	LogID         string `json:"logId,omitempty"`
	ModelName     string `json:"modelName,omitempty"`
	ContentOffset int    `json:"contentOffset"`
	StartedAt     string `json:"startedAt"`
	FinishedAt    string `json:"finishedAt,omitempty"`
}
