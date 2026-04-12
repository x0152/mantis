package types

import "time"

type PlanRun struct {
	ID         string        `json:"id"`
	PlanID     string        `json:"planId"`
	Status     string        `json:"status"`
	Trigger    string        `json:"trigger"`
	Steps      []PlanStepRun `json:"steps"`
	StartedAt  time.Time     `json:"startedAt"`
	FinishedAt *time.Time    `json:"finishedAt,omitempty"`
}

type PlanStepRun struct {
	NodeID     string     `json:"nodeId"`
	Status     string     `json:"status"`
	Result     string     `json:"result,omitempty"`
	MessageID  string     `json:"messageId,omitempty"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
}
