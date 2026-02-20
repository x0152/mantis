package chat

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	"mantis/apps/chat/api"
	usecases "mantis/apps/chat/use_cases"
	"mantis/core/agents"
	artifactplugin "mantis/core/plugins/artifact"
	modelplugin "mantis/core/plugins/model"
	"mantis/core/protocols"
	"mantis/core/types"
	messageworkflow "mantis/core/workflows/message"
	"mantis/shared"
)

type App struct {
	endpoints *api.Endpoints
}

func NewApp(
	sessionStore protocols.Store[string, types.ChatSession],
	messageStore protocols.Store[string, types.ChatMessage],
	modelStore protocols.Store[string, types.Model],
	channelStore protocols.Store[string, types.Channel],
	configStore protocols.Store[string, types.Config],
	mantisAgent *agents.MantisAgent,
	buf *shared.Buffer,
	artifactMgr *artifactplugin.Manager,
) *App {
	modelResolver := modelplugin.NewResolver(channelStore, configStore)
	workflow := messageworkflow.New(messageStore, modelStore, mantisAgent, buf, modelResolver, artifactMgr)
	return &App{
		endpoints: api.NewEndpoints(api.UseCases{
			GetCurrentSession: usecases.NewGetCurrentSession(sessionStore),
			ResetContext:      usecases.NewResetContext(sessionStore),
			ListMessages:      usecases.NewListMessages(messageStore, buf),
			SendMessage:       usecases.NewSendMessage(workflow),
			ClearHistory:      usecases.NewClearHistory(sessionStore, messageStore),
		}),
	}
}

func (a *App) Register(api huma.API) {
	a.endpoints.Register(api)
}

func (a *App) Handler() http.Handler {
	r := chi.NewMux()
	a.Register(humachi.New(r, huma.DefaultConfig("Mantis Chat API", "1.0.0")))
	return r
}
