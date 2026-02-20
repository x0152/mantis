package telegram

import (
	"context"

	usecases "mantis/apps/telegram/use_cases"
)

func (a *App) getOrCreateSession(ctx context.Context, chatID string) string {
	if a.ucSession == nil {
		return ""
	}
	out, err := a.ucSession.Execute(ctx, usecases.SessionModeGetOrCreate)
	if err != nil {
		return ""
	}
	return out
}

func (a *App) resetSession(ctx context.Context, chatID string) (string, error) {
	if a.ucSession == nil {
		return "", nil
	}
	return a.ucSession.Execute(ctx, usecases.SessionModeReset)
}
