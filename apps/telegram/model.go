package telegram

import (
	"context"
	adapter "mantis/infrastructure/adapters/channel"
)

func (a *App) handleModelCommand(ctx context.Context, channelID string, args string) (adapter.Reply, error) {
	if a.ucModelCommand == nil {
		return adapter.Reply{}, nil
	}
	return a.ucModelCommand.Execute(ctx, channelID, args)
}
