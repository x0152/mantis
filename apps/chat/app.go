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
	"mantis/core/plugins/pipeline"
	"mantis/core/protocols"
	"mantis/core/types"
	messageworkflow "mantis/core/workflows/message"
	"mantis/shared"
)

type App struct {
	endpoints *api.Endpoints
	workflow  *messageworkflow.Workflow
}

func NewApp(
	sessionStore protocols.Store[string, types.ChatSession],
	messageStore protocols.Store[string, types.ChatMessage],
	modelStore protocols.Store[string, types.Model],
	presetStore protocols.Store[string, types.Preset],
	channelStore protocols.Store[string, types.Channel],
	settingsStore protocols.Store[string, types.Settings],
	mantisAgent *agents.MantisAgent,
	buf *shared.Buffer,
	artifactMgr *artifactplugin.Manager,
	memoryExtractor pipeline.MemoryExtractor,
	cancellations *pipeline.Cancellations,
	planRunner protocols.PlanRunner,
) *App {
	modelResolver := modelplugin.NewResolver(channelStore, settingsStore, presetStore)
	workflow := messageworkflow.New(messageStore, modelStore, mantisAgent, buf, modelResolver, artifactMgr, memoryExtractor, cancellations)
	return &App{
		workflow: workflow,
		endpoints: api.NewEndpoints(api.UseCases{
			GetCurrentSession: usecases.NewGetCurrentSession(sessionStore),
			ResetContext:      usecases.NewResetContext(sessionStore),
			ListSessions:      usecases.NewListSessions(sessionStore, buf),
			CreateSession:     usecases.NewCreateSession(sessionStore),
			UpdateSession:     usecases.NewUpdateSession(sessionStore),
			DeleteSession:     usecases.NewDeleteSession(sessionStore, messageStore),
			ListMessages:      usecases.NewListMessages(messageStore, buf),
			SendMessage:       usecases.NewSendMessage(workflow, sessionStore, mantisAgent.Limits()),
			ClearHistory:      usecases.NewClearHistory(sessionStore, messageStore),
			StopGeneration:    usecases.NewStopGeneration(cancellations, planRunner),
		}),
	}
}

func (a *App) SetAttachmentDir(dir string) {
	a.workflow.SetAttachmentDir(dir)
}

func (a *App) Register(api huma.API) {
	a.endpoints.Register(api)
}

func (a *App) Handler() http.Handler {
	r := chi.NewMux()
	a.Register(humachi.New(r, huma.DefaultConfig("Mantis Chat API", "1.0.0")))
	return r
}
