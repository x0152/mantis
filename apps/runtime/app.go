package runtime

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"mantis/apps/runtime/api"
	"mantis/apps/runtime/keys"
	"mantis/apps/runtime/spec"
	"mantis/core/protocols"
	"mantis/core/types"
)

type App struct {
	endpoints *api.Endpoints
}

func NewApp(
	rt protocols.Runtime,
	connectionStore protocols.Store[string, types.Connection],
	keyIssuer *keys.Issuer,
	specBuilder *spec.Builder,
	token string,
) *App {
	return &App{endpoints: api.NewEndpoints(rt, connectionStore, keyIssuer, specBuilder, token)}
}

func (a *App) Mount(r chi.Router) {
	a.endpoints.Mount(r)
}

func (a *App) Handler() http.Handler {
	r := chi.NewMux()
	a.Mount(r)
	return r
}
