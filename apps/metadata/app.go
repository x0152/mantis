package metadata

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	"mantis/apps/metadata/api"
	usecases "mantis/apps/metadata/use_cases"
	"mantis/core/protocols"
	"mantis/core/types"
)

type App struct {
	endpoints *api.Endpoints
}

func NewApp(
	configStore protocols.Store[string, types.Config],
	llmConnStore protocols.Store[string, types.LlmConnection],
	modelStore protocols.Store[string, types.Model],
	connectionStore protocols.Store[string, types.Connection],
	cronJobStore protocols.Store[string, types.CronJob],
	guardProfileStore protocols.Store[string, types.GuardProfile],
	channelStore protocols.Store[string, types.Channel],
) *App {
	return &App{
		endpoints: api.NewEndpoints(api.UseCases{
			GetConfig:        usecases.NewGetConfig(configStore),
			UpdateConfig:     usecases.NewUpdateConfig(configStore),
			CreateLlmConn:    usecases.NewCreateLlmConnection(llmConnStore),
			GetLlmConn:       usecases.NewGetLlmConnection(llmConnStore),
			ListLlmConns:     usecases.NewListLlmConnections(llmConnStore),
			UpdateLlmConn:    usecases.NewUpdateLlmConnection(llmConnStore),
			DeleteLlmConn:    usecases.NewDeleteLlmConnection(llmConnStore),
			CreateModel:      usecases.NewCreateModel(modelStore),
			GetModel:         usecases.NewGetModel(modelStore),
			ListModels:       usecases.NewListModels(modelStore),
			UpdateModel:      usecases.NewUpdateModel(modelStore),
			DeleteModel:      usecases.NewDeleteModel(modelStore),
			CreateConnection: usecases.NewCreateConnection(connectionStore),
			GetConnection:    usecases.NewGetConnection(connectionStore),
			ListConnections:  usecases.NewListConnections(connectionStore),
			UpdateConnection: usecases.NewUpdateConnection(connectionStore),
			DeleteConnection: usecases.NewDeleteConnection(connectionStore),
			AddMemory:        usecases.NewAddMemory(connectionStore),
			DeleteMemory:     usecases.NewDeleteMemory(connectionStore),
			CreateCronJob:    usecases.NewCreateCronJob(cronJobStore),
			GetCronJob:       usecases.NewGetCronJob(cronJobStore),
			ListCronJobs:     usecases.NewListCronJobs(cronJobStore),
			UpdateCronJob:    usecases.NewUpdateCronJob(cronJobStore),
			DeleteCronJob:    usecases.NewDeleteCronJob(cronJobStore),
			CreateGuardProfile:  usecases.NewCreateGuardProfile(guardProfileStore),
			ListGuardProfiles:   usecases.NewListGuardProfiles(guardProfileStore),
			UpdateGuardProfile:  usecases.NewUpdateGuardProfile(guardProfileStore),
			DeleteGuardProfile:  usecases.NewDeleteGuardProfile(guardProfileStore),
			CreateChannel:    usecases.NewCreateChannel(channelStore),
			GetChannel:       usecases.NewGetChannel(channelStore),
			ListChannels:     usecases.NewListChannels(channelStore),
			UpdateChannel:    usecases.NewUpdateChannel(channelStore),
			DeleteChannel:    usecases.NewDeleteChannel(channelStore),
		}),
	}
}

func (a *App) Register(api huma.API) {
	a.endpoints.Register(api)
}

func (a *App) Handler() http.Handler {
	r := chi.NewMux()
	a.Register(humachi.New(r, huma.DefaultConfig("Mantis Metadata API", "1.0.0")))
	return r
}
