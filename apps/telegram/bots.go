package telegram

import (
	"context"
)

func (a *App) stopBots() {
	if a.ucSyncBots != nil {
		a.ucSyncBots.Stop()
	}
}

func (a *App) syncBots(ctx context.Context) {
	if a.ucSyncBots != nil {
		a.ucSyncBots.Execute(ctx)
	}
}
