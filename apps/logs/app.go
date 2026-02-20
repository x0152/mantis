package logs

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	"mantis/apps/logs/api"
	usecases "mantis/apps/logs/use_cases"
	"mantis/core/protocols"
	"mantis/core/types"
)

type App struct {
	endpoints *api.Endpoints
}

func NewApp(logStore protocols.Store[string, types.SessionLog]) *App {
	return &App{
		endpoints: api.NewEndpoints(api.UseCases{
			ListSessionLogs: usecases.NewListSessionLogs(logStore),
			GetSessionLog:   usecases.NewGetSessionLog(logStore),
			ClearLogs:       usecases.NewClearLogs(logStore),
		}),
	}
}

func (a *App) Register(api huma.API) {
	a.endpoints.Register(api)
}

func (a *App) Handler() http.Handler {
	r := chi.NewMux()
	a.Register(humachi.New(r, huma.DefaultConfig("Mantis Logs API", "1.0.0")))
	return r
}
