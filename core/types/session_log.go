package types

import "time"

type SessionLog struct {
	ID           string     `json:"id"`
	ConnectionID string     `json:"connectionId"`
	AgentName    string     `json:"agentName"`
	Prompt       string     `json:"prompt,omitempty"`
	Status       string     `json:"status"`
	MessageID    string     `json:"messageId,omitempty"`
	StepID       string     `json:"stepId,omitempty"`
	ModelName    string     `json:"modelName,omitempty"`
	Entries      []LogEntry `json:"entries"`
	StartedAt    time.Time  `json:"startedAt"`
	FinishedAt   *time.Time `json:"finishedAt,omitempty"`
}

type LogEntry struct {
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
