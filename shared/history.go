package shared

import (
	"context"
	"sort"
	"strings"
	"time"

	"mantis/core/protocols"
	"mantis/core/types"
)

func BuildHistory(ctx context.Context, store protocols.Store[string, types.ChatMessage], sessionID string) ([]protocols.LLMMessage, error) {
	all, err := store.List(ctx, types.ListQuery{})
	if err != nil {
		return nil, err
	}

	includeCron := !strings.HasPrefix(sessionID, "cron:")
	cronCutoff := time.Now().Add(-24 * time.Hour)

	var session []types.ChatMessage
	for _, m := range all {
		if m.Status != "" {
			continue
		}
		if m.SessionID == sessionID {
			session = append(session, m)
			continue
		}
		if includeCron && strings.HasPrefix(m.SessionID, "cron:") && m.CreatedAt.After(cronCutoff) {
			session = append(session, m)
		}
	}
	sort.Slice(session, func(i, j int) bool {
		return session[i].CreatedAt.Before(session[j].CreatedAt)
	})

	msgs := make([]protocols.LLMMessage, 0, len(session))
	for _, m := range session {
		content := m.Content
		if m.SessionID != sessionID {
			if m.Role == "user" {
				content = "[Scheduled task] " + content
			} else {
				content = "[Scheduled task result] " + content
			}
		}
		msgs = append(msgs, protocols.LLMMessage{Role: m.Role, Content: content})
	}
	return msgs, nil
}
