package shared

import (
	"context"
	"sort"
	"strings"
	"time"

	"mantis/core/protocols"
	"mantis/core/types"
)

func BuildHistory(
	ctx context.Context,
	messageStore protocols.Store[string, types.ChatMessage],
	sessionStore protocols.Store[string, types.ChatSession],
	sessionID string,
) ([]protocols.LLMMessage, error) {
	all, err := messageStore.List(ctx, types.ListQuery{})
	if err != nil {
		return nil, err
	}

	var session types.ChatSession
	if sessionStore != nil && sessionID != "" {
		if sessions, err := sessionStore.Get(ctx, []string{sessionID}); err == nil {
			if s, ok := sessions[sessionID]; ok {
				session = s
			}
		}
	}

	includePlans := !strings.HasPrefix(sessionID, "plan:")
	planCutoff := time.Now().Add(-24 * time.Hour)

	var kept []types.ChatMessage
	for _, m := range all {
		if m.Status != "" {
			continue
		}
		if m.SessionID == sessionID {
			if session.SummarizedUpTo != nil && !m.CreatedAt.After(*session.SummarizedUpTo) {
				continue
			}
			kept = append(kept, m)
			continue
		}
		if includePlans && strings.HasPrefix(m.SessionID, "plan:") && m.CreatedAt.After(planCutoff) {
			kept = append(kept, m)
		}
	}
	sort.Slice(kept, func(i, j int) bool {
		return kept[i].CreatedAt.Before(kept[j].CreatedAt)
	})

	msgs := make([]protocols.LLMMessage, 0, len(kept)+1)
	if session.SummaryText != "" {
		msgs = append(msgs, protocols.LLMMessage{
			Role:    "system",
			Content: "[Prior conversation summary]\n" + session.SummaryText,
		})
	}
	for _, m := range kept {
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
