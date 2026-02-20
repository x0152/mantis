package mappers

import (
	"encoding/json"

	"mantis/core/types"
	"mantis/infrastructure/models"
)

func SessionLogToRow(s types.SessionLog) models.SessionLogRow {
	entries, _ := json.Marshal(s.Entries)
	return models.SessionLogRow{
		ID: s.ID, ConnectionID: s.ConnectionID, AgentName: s.AgentName,
		Prompt: s.Prompt, Status: s.Status,
		MessageID: s.MessageID, StepID: s.StepID, ModelName: s.ModelName,
		Entries: entries, StartedAt: s.StartedAt, FinishedAt: s.FinishedAt,
	}
}

func SessionLogFromRow(r models.SessionLogRow) types.SessionLog {
	var entries []types.LogEntry
	_ = json.Unmarshal(r.Entries, &entries)
	if entries == nil {
		entries = []types.LogEntry{}
	}
	return types.SessionLog{
		ID: r.ID, ConnectionID: r.ConnectionID, AgentName: r.AgentName,
		Prompt: r.Prompt, Status: r.Status,
		MessageID: r.MessageID, StepID: r.StepID, ModelName: r.ModelName,
		Entries: entries, StartedAt: r.StartedAt, FinishedAt: r.FinishedAt,
	}
}
