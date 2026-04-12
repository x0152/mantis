package metadata

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	"mantis/apps/metadata/api"
	usecases "mantis/apps/metadata/use_cases"
	"mantis/apps/plans"
	"mantis/core/protocols"
	"mantis/core/types"
)

type App struct {
	endpoints *api.Endpoints
}

func NewApp(
	settingsStore protocols.Store[string, types.Settings],
	llmConnStore protocols.Store[string, types.LlmConnection],
	modelStore protocols.Store[string, types.Model],
	presetStore protocols.Store[string, types.Preset],
	connectionStore protocols.Store[string, types.Connection],
	skillStore protocols.Store[string, types.Skill],
	planStore protocols.Store[string, types.Plan],
	runStore protocols.Store[string, types.PlanRun],
	planRunner *plans.Runner,
	cronJobStore protocols.Store[string, types.CronJob],
	guardProfileStore protocols.Store[string, types.GuardProfile],
	channelStore protocols.Store[string, types.Channel],
) *App {
	return &App{
		endpoints: api.NewEndpoints(api.UseCases{
			GetSettings:        usecases.NewGetSettings(settingsStore),
			UpdateSettings:     usecases.NewUpdateSettings(settingsStore),
			CreateLlmConn:      usecases.NewCreateLlmConnection(llmConnStore),
			GetLlmConn:         usecases.NewGetLlmConnection(llmConnStore),
			ListLlmConns:       usecases.NewListLlmConnections(llmConnStore),
			UpdateLlmConn:      usecases.NewUpdateLlmConnection(llmConnStore),
			DeleteLlmConn:      usecases.NewDeleteLlmConnection(llmConnStore),
			CreateModel:        usecases.NewCreateModel(modelStore),
			GetModel:           usecases.NewGetModel(modelStore),
			ListModels:         usecases.NewListModels(modelStore),
			UpdateModel:        usecases.NewUpdateModel(modelStore),
			DeleteModel:        usecases.NewDeleteModel(modelStore),
			CreatePreset:       usecases.NewCreatePreset(presetStore),
			ListPresets:        usecases.NewListPresets(presetStore),
			UpdatePreset:       usecases.NewUpdatePreset(presetStore),
			DeletePreset:       usecases.NewDeletePreset(presetStore),
			CreateConnection:   usecases.NewCreateConnection(connectionStore),
			GetConnection:      usecases.NewGetConnection(connectionStore),
			ListConnections:    usecases.NewListConnections(connectionStore),
			UpdateConnection:   usecases.NewUpdateConnection(connectionStore),
			DeleteConnection:   usecases.NewDeleteConnection(connectionStore),
			CreateSkill:        usecases.NewCreateSkill(skillStore),
			ListSkills:         usecases.NewListSkills(skillStore),
			UpdateSkill:        usecases.NewUpdateSkill(skillStore),
			DeleteSkill:        usecases.NewDeleteSkill(skillStore),
			AddMemory:          usecases.NewAddMemory(connectionStore),
			DeleteMemory:       usecases.NewDeleteMemory(connectionStore),
			CreatePlan:         usecases.NewCreatePlan(planStore),
			ListPlans:          usecases.NewListPlans(planStore),
			UpdatePlan:         usecases.NewUpdatePlan(planStore),
			DeletePlan:         usecases.NewDeletePlan(planStore),
			ListPlanRuns:       usecases.NewListPlanRuns(runStore),
			GetPlanRun:         usecases.NewGetPlanRun(runStore),
			PlanRunner:         planRunner,
			CreateCronJob:      usecases.NewCreateCronJob(cronJobStore),
			GetCronJob:         usecases.NewGetCronJob(cronJobStore),
			ListCronJobs:       usecases.NewListCronJobs(cronJobStore),
			UpdateCronJob:      usecases.NewUpdateCronJob(cronJobStore),
			DeleteCronJob:      usecases.NewDeleteCronJob(cronJobStore),
			CreateGuardProfile: usecases.NewCreateGuardProfile(guardProfileStore),
			ListGuardProfiles:  usecases.NewListGuardProfiles(guardProfileStore),
			UpdateGuardProfile: usecases.NewUpdateGuardProfile(guardProfileStore),
			DeleteGuardProfile: usecases.NewDeleteGuardProfile(guardProfileStore),
			CreateChannel:      usecases.NewCreateChannel(channelStore),
			GetChannel:         usecases.NewGetChannel(channelStore),
			ListChannels:       usecases.NewListChannels(channelStore),
			UpdateChannel:      usecases.NewUpdateChannel(channelStore),
			DeleteChannel:      usecases.NewDeleteChannel(channelStore),
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
