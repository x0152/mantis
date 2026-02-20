package telegram

import (
	"context"
	"time"

	usecases "mantis/apps/telegram/use_cases"
	"mantis/core/agents"
	artifactplugin "mantis/core/plugins/artifact"
	modelplugin "mantis/core/plugins/model"
	sessionplugin "mantis/core/plugins/session"
	"mantis/core/protocols"
	"mantis/core/types"
	messageworkflow "mantis/core/workflows/message"
	"mantis/shared"
)

type App struct {
	ucSession       *usecases.Session
	ucModelCommand  *usecases.HandleModelCommand
	ucHandleMessage *usecases.HandleMessage
	ucSyncBots      *usecases.SyncBots
	syncFreq        time.Duration
}

func NewApp(
	channelStore protocols.Store[string, types.Channel],
	sessionStore protocols.Store[string, types.ChatSession],
	messageStore protocols.Store[string, types.ChatMessage],
	modelStore protocols.Store[string, types.Model],
	agent *agents.MantisAgent,
	buffer *shared.Buffer,
	artifactMgr *artifactplugin.Manager,
	asr protocols.ASR,
	tts protocols.TTS,
) *App {
	if artifactMgr == nil {
		artifactMgr = artifactplugin.NewManager(nil)
	}
	modelResolver := modelplugin.NewResolver(channelStore, nil)
	workflow := messageworkflow.New(messageStore, modelStore, agent, buffer, modelResolver, artifactMgr)

	sessionUC := usecases.NewSession(sessionplugin.NewPolicy(sessionStore))
	modelCommandUC := usecases.NewHandleModelCommand(modelStore, channelStore)
	handleMessageUC := usecases.NewHandleMessage(sessionUC, modelCommandUC, channelStore, messageStore, workflow, buffer, asr, tts)

	app := &App{
		ucSession:       sessionUC,
		ucModelCommand:  modelCommandUC,
		ucHandleMessage: handleMessageUC,
		syncFreq:        30 * time.Second,
	}
	app.ucSyncBots = usecases.NewSyncBots(channelStore, app.makeHandler)
	return app
}

func (a *App) Start(ctx context.Context) {
	if a.ucSyncBots == nil {
		return
	}

	a.syncBots(ctx)
	ticker := time.NewTicker(a.syncFreq)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			a.stopBots()
			return
		case <-ticker.C:
			a.syncBots(ctx)
		}
	}
}
