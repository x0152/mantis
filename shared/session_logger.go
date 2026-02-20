package shared

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"mantis/core/protocols"
	"mantis/core/types"
)

type ctxKey int

const (
	ctxKeyStepID    ctxKey = iota
	ctxKeyMessageID ctxKey = iota
	ctxKeyLogHolder ctxKey = iota
)

type ToolMeta struct {
	LogID     string
	ModelName string
}

func ContextWithStep(ctx context.Context, stepID, messageID string) context.Context {
	ctx = context.WithValue(ctx, ctxKeyStepID, stepID)
	ctx = context.WithValue(ctx, ctxKeyMessageID, messageID)
	ctx = context.WithValue(ctx, ctxKeyLogHolder, &ToolMeta{})
	return ctx
}

func StepFromContext(ctx context.Context) (stepID, messageID string) {
	if v, ok := ctx.Value(ctxKeyStepID).(string); ok {
		stepID = v
	}
	if v, ok := ctx.Value(ctxKeyMessageID).(string); ok {
		messageID = v
	}
	return
}

func ToolMetaFromContext(ctx context.Context) *ToolMeta {
	if h, ok := ctx.Value(ctxKeyLogHolder).(*ToolMeta); ok {
		return h
	}
	return nil
}

func SetLogID(ctx context.Context, logID string) {
	if m := ToolMetaFromContext(ctx); m != nil {
		m.LogID = logID
	}
}

func GetLogID(ctx context.Context) string {
	if m := ToolMetaFromContext(ctx); m != nil {
		return m.LogID
	}
	return ""
}

func SetModelName(ctx context.Context, name string) {
	if m := ToolMetaFromContext(ctx); m != nil {
		m.ModelName = name
	}
}

func GetModelName(ctx context.Context) string {
	if m := ToolMetaFromContext(ctx); m != nil {
		return m.ModelName
	}
	return ""
}

type SessionLogger struct {
	store protocols.Store[string, types.SessionLog]
}

func NewSessionLogger(store protocols.Store[string, types.SessionLog]) *SessionLogger {
	return &SessionLogger{store: store}
}

func (l *SessionLogger) Wrap(ctx context.Context, connectionID, agentName, prompt string, src <-chan types.StreamEvent) <-chan types.StreamEvent {
	stepID, messageID := StepFromContext(ctx)

	now := time.Now().UTC()
	session := types.SessionLog{
		ID:           uuid.New().String(),
		ConnectionID: connectionID,
		AgentName:    agentName,
		Prompt:       prompt,
		MessageID:    messageID,
		StepID:       stepID,
		ModelName:    GetModelName(ctx),
		Status:       "running",
		Entries:      []types.LogEntry{},
		StartedAt:    now,
	}
	if _, err := l.store.Create(ctx, []types.SessionLog{session}); err != nil {
		log.Printf("session_logger: failed to create log: %v", err)
		return src
	}
	SetLogID(ctx, session.ID)

	out := make(chan types.StreamEvent, 32)
	go func() {
		defer close(out)
		var entries []types.LogEntry
		var textBuf strings.Builder

		save := func() {
			session.Entries = entries
			if _, err := l.store.Update(context.Background(), []types.SessionLog{session}); err != nil {
				log.Printf("session_logger: failed to update log: %v", err)
			}
		}

		flushText := func() {
			if textBuf.Len() > 0 {
				entries = append(entries, types.LogEntry{
					Type: "thought", Content: strings.TrimSpace(textBuf.String()), Timestamp: time.Now().UTC(),
				})
				textBuf.Reset()
			}
		}

		for event := range src {
			out <- event

			switch event.Type {
			case "text":
				textBuf.WriteString(event.Delta)
			case "thinking":
				flushText()
				entries = append(entries, types.LogEntry{Type: "thought", Content: strings.TrimSpace(event.Delta), Timestamp: time.Now().UTC()})
				save()
			case "tool_start":
				flushText()
				entries = append(entries, types.LogEntry{Type: "command", Content: event.Delta, Timestamp: time.Now().UTC()})
				save()
			case "tool_end":
				entries = append(entries, types.LogEntry{Type: "output", Content: event.Delta, Timestamp: time.Now().UTC()})
				save()
			case "error":
				flushText()
				entries = append(entries, types.LogEntry{Type: "error", Content: event.Delta, Timestamp: time.Now().UTC()})
				save()
			}
		}
		flushText()

		finished := time.Now().UTC()
		session.Status = "finished"
		session.FinishedAt = &finished
		save()
	}()

	return out
}
