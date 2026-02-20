package telegram

import (
	"context"

	adapter "mantis/infrastructure/adapters/channel"
)

func (a *App) makeHandler(channelID string) adapter.MessageHandler {
	if a.ucHandleMessage == nil {
		return func(_ context.Context, _ string, _ string, _ []adapter.FileAttachment) (adapter.Reply, error) {
			return adapter.Reply{}, nil
		}
	}
	return a.ucHandleMessage.Handler(channelID)
}
