package auth

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	"mantis/apps/auth/api"
	usecases "mantis/apps/auth/use_cases"
	"mantis/core/protocols"
	"mantis/core/types"
)

type App struct {
	endpoints *api.Endpoints
	bootstrap *usecases.Bootstrap
}

func NewApp(userStore protocols.Store[string, types.User]) *App {
	return &App{
		endpoints: api.NewEndpoints(api.UseCases{
			Login: usecases.NewLogin(userStore),
			Me:    usecases.NewMe(userStore),
		}),
		bootstrap: usecases.NewBootstrap(userStore),
	}
}

func (a *App) Register(api huma.API) {
	a.endpoints.Register(api)
}

func (a *App) Bootstrap(ctx context.Context, name, token string) (types.User, error) {
	return a.bootstrap.Execute(ctx, name, token)
}
