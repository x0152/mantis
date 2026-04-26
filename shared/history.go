package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"mantis/core/protocols"
	"mantis/core/types"
)

const (
	historyStepArgsMax = 240
	historyMaxSteps    = 25
)

func stepsSummary(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var steps []types.Step
	if err := json.Unmarshal(raw, &steps); err != nil || len(steps) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\n[Tool activity from this turn — DO NOT redo these calls; their results are already known:")
	shown := len(steps)
	if shown > historyMaxSteps {
		shown = historyMaxSteps
	}
	for _, s := range steps[:shown] {
		status := strings.TrimSpace(s.Status)
		if status == "" {
			status = "done"
		}
		tool := s.Tool
		if tool == "" {
			tool = "tool"
		}
		sb.WriteString(fmt.Sprintf("\n- %s [%s]", tool, status))
		if label := strings.TrimSpace(s.Label); label != "" && label != tool {
			sb.WriteString(" — " + label)
		}
		if args := strings.TrimSpace(s.Args); args != "" {
			sb.WriteString("\n  args: " + truncate(args, historyStepArgsMax))
		}
		if result := strings.TrimSpace(s.Result); result != "" {
			sb.WriteString("\n  result: " + result)
		}
	}
	if extra := len(steps) - shown; extra > 0 {
		sb.WriteString(fmt.Sprintf("\n- …and %d more tool call(s) omitted", extra))
	}
	sb.WriteString("\n]")
	return sb.String()
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…[truncated]"
}

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
		if m.Role == "assistant" {
			if summary := stepsSummary(m.Steps); summary != "" {
				content += summary
			}
		}
		msgs = append(msgs, protocols.LLMMessage{Role: m.Role, Content: content})
	}
	return msgs, nil
}
