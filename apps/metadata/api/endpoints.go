package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "mantis/apps/metadata/use_cases"
	"mantis/core/base"
)

type UseCases struct {
	GetConfig        *usecases.GetConfig
	UpdateConfig     *usecases.UpdateConfig
	CreateLlmConn    *usecases.CreateLlmConnection
	GetLlmConn       *usecases.GetLlmConnection
	ListLlmConns     *usecases.ListLlmConnections
	UpdateLlmConn    *usecases.UpdateLlmConnection
	DeleteLlmConn    *usecases.DeleteLlmConnection
	CreateModel      *usecases.CreateModel
	GetModel         *usecases.GetModel
	ListModels       *usecases.ListModels
	UpdateModel      *usecases.UpdateModel
	DeleteModel      *usecases.DeleteModel
	CreateConnection *usecases.CreateConnection
	GetConnection    *usecases.GetConnection
	ListConnections  *usecases.ListConnections
	UpdateConnection *usecases.UpdateConnection
	DeleteConnection *usecases.DeleteConnection
	AddMemory        *usecases.AddMemory
	DeleteMemory     *usecases.DeleteMemory
	CreateCronJob    *usecases.CreateCronJob
	GetCronJob       *usecases.GetCronJob
	ListCronJobs     *usecases.ListCronJobs
	UpdateCronJob    *usecases.UpdateCronJob
	DeleteCronJob    *usecases.DeleteCronJob
	CreateGuardProfile  *usecases.CreateGuardProfile
	ListGuardProfiles   *usecases.ListGuardProfiles
	UpdateGuardProfile  *usecases.UpdateGuardProfile
	DeleteGuardProfile  *usecases.DeleteGuardProfile
	CreateChannel    *usecases.CreateChannel
	GetChannel       *usecases.GetChannel
	ListChannels     *usecases.ListChannels
	UpdateChannel    *usecases.UpdateChannel
	DeleteChannel    *usecases.DeleteChannel
}

type Endpoints struct {
	uc UseCases
}

func NewEndpoints(uc UseCases) *Endpoints {
	return &Endpoints{uc: uc}
}

func (e *Endpoints) Register(api huma.API) {
	huma.Register(api, huma.Operation{OperationID: "get-config", Method: http.MethodGet, Path: "/api/config"}, e.getConfig)
	huma.Register(api, huma.Operation{OperationID: "update-config", Method: http.MethodPut, Path: "/api/config"}, e.updateConfig)

	huma.Register(api, huma.Operation{OperationID: "create-llm-connection", Method: http.MethodPost, Path: "/api/llm-connections", DefaultStatus: 201}, e.createLlmConnection)
	huma.Register(api, huma.Operation{OperationID: "list-llm-connections", Method: http.MethodGet, Path: "/api/llm-connections"}, e.listLlmConnections)
	huma.Register(api, huma.Operation{OperationID: "get-llm-connection", Method: http.MethodGet, Path: "/api/llm-connections/{id}"}, e.getLlmConnection)
	huma.Register(api, huma.Operation{OperationID: "update-llm-connection", Method: http.MethodPut, Path: "/api/llm-connections/{id}"}, e.updateLlmConnection)
	huma.Register(api, huma.Operation{OperationID: "delete-llm-connection", Method: http.MethodDelete, Path: "/api/llm-connections/{id}", DefaultStatus: 204}, e.deleteLlmConnection)

	huma.Register(api, huma.Operation{OperationID: "create-model", Method: http.MethodPost, Path: "/api/models", DefaultStatus: 201}, e.createModel)
	huma.Register(api, huma.Operation{OperationID: "list-models", Method: http.MethodGet, Path: "/api/models"}, e.listModels)
	huma.Register(api, huma.Operation{OperationID: "get-model", Method: http.MethodGet, Path: "/api/models/{id}"}, e.getModel)
	huma.Register(api, huma.Operation{OperationID: "update-model", Method: http.MethodPut, Path: "/api/models/{id}"}, e.updateModel)
	huma.Register(api, huma.Operation{OperationID: "delete-model", Method: http.MethodDelete, Path: "/api/models/{id}", DefaultStatus: 204}, e.deleteModel)

	huma.Register(api, huma.Operation{OperationID: "create-connection", Method: http.MethodPost, Path: "/api/connections", DefaultStatus: 201}, e.createConnection)
	huma.Register(api, huma.Operation{OperationID: "list-connections", Method: http.MethodGet, Path: "/api/connections"}, e.listConnections)
	huma.Register(api, huma.Operation{OperationID: "get-connection", Method: http.MethodGet, Path: "/api/connections/{id}"}, e.getConnection)
	huma.Register(api, huma.Operation{OperationID: "update-connection", Method: http.MethodPut, Path: "/api/connections/{id}"}, e.updateConnection)
	huma.Register(api, huma.Operation{OperationID: "delete-connection", Method: http.MethodDelete, Path: "/api/connections/{id}", DefaultStatus: 204}, e.deleteConnection)
	huma.Register(api, huma.Operation{OperationID: "add-memory", Method: http.MethodPost, Path: "/api/connections/{id}/memories", DefaultStatus: 201}, e.addMemory)
	huma.Register(api, huma.Operation{OperationID: "delete-memory", Method: http.MethodDelete, Path: "/api/connections/{id}/memories/{memoryId}", DefaultStatus: 204}, e.deleteMemory)

	huma.Register(api, huma.Operation{OperationID: "create-cron-job", Method: http.MethodPost, Path: "/api/cron-jobs", DefaultStatus: 201}, e.createCronJob)
	huma.Register(api, huma.Operation{OperationID: "list-cron-jobs", Method: http.MethodGet, Path: "/api/cron-jobs"}, e.listCronJobs)
	huma.Register(api, huma.Operation{OperationID: "get-cron-job", Method: http.MethodGet, Path: "/api/cron-jobs/{id}"}, e.getCronJob)
	huma.Register(api, huma.Operation{OperationID: "update-cron-job", Method: http.MethodPut, Path: "/api/cron-jobs/{id}"}, e.updateCronJob)
	huma.Register(api, huma.Operation{OperationID: "delete-cron-job", Method: http.MethodDelete, Path: "/api/cron-jobs/{id}", DefaultStatus: 204}, e.deleteCronJob)

	huma.Register(api, huma.Operation{OperationID: "create-guard-profile", Method: http.MethodPost, Path: "/api/guard-profiles", DefaultStatus: 201}, e.createGuardProfile)
	huma.Register(api, huma.Operation{OperationID: "list-guard-profiles", Method: http.MethodGet, Path: "/api/guard-profiles"}, e.listGuardProfiles)
	huma.Register(api, huma.Operation{OperationID: "update-guard-profile", Method: http.MethodPut, Path: "/api/guard-profiles/{id}"}, e.updateGuardProfile)
	huma.Register(api, huma.Operation{OperationID: "delete-guard-profile", Method: http.MethodDelete, Path: "/api/guard-profiles/{id}", DefaultStatus: 204}, e.deleteGuardProfile)

	huma.Register(api, huma.Operation{OperationID: "create-channel", Method: http.MethodPost, Path: "/api/channels", DefaultStatus: 201}, e.createChannel)
	huma.Register(api, huma.Operation{OperationID: "list-channels", Method: http.MethodGet, Path: "/api/channels"}, e.listChannels)
	huma.Register(api, huma.Operation{OperationID: "get-channel", Method: http.MethodGet, Path: "/api/channels/{id}"}, e.getChannel)
	huma.Register(api, huma.Operation{OperationID: "update-channel", Method: http.MethodPut, Path: "/api/channels/{id}"}, e.updateChannel)
	huma.Register(api, huma.Operation{OperationID: "delete-channel", Method: http.MethodDelete, Path: "/api/channels/{id}", DefaultStatus: 204}, e.deleteChannel)
}

func (e *Endpoints) getConfig(ctx context.Context, _ *struct{}) (*ConfigOutput, error) {
	cfg, err := e.uc.GetConfig.Execute(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return toConfigOutput(cfg), nil
}

func (e *Endpoints) updateConfig(ctx context.Context, input *UpdateConfigInput) (*ConfigOutput, error) {
	cfg, err := e.uc.UpdateConfig.Execute(ctx, configDataFromInput(input))
	if err != nil {
		return nil, mapErr(err)
	}
	return toConfigOutput(cfg), nil
}

func (e *Endpoints) createLlmConnection(ctx context.Context, input *CreateLlmConnectionInput) (*LlmConnectionOutput, error) {
	id, provider, baseURL, apiKey := llmConnectionFromCreateInput(input)
	c, err := e.uc.CreateLlmConn.Execute(ctx, id, provider, baseURL, apiKey)
	if err != nil {
		return nil, mapErr(err)
	}
	return toLlmConnectionOutput(c), nil
}

func (e *Endpoints) listLlmConnections(ctx context.Context, _ *struct{}) (*LlmConnectionsOutput, error) {
	items, err := e.uc.ListLlmConns.Execute(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return toLlmConnectionsOutput(items), nil
}

func (e *Endpoints) getLlmConnection(ctx context.Context, input *LlmConnectionIDInput) (*LlmConnectionOutput, error) {
	c, err := e.uc.GetLlmConn.Execute(ctx, input.ID)
	if err != nil {
		return nil, mapErr(err)
	}
	return toLlmConnectionOutput(c), nil
}

func (e *Endpoints) updateLlmConnection(ctx context.Context, input *UpdateLlmConnectionInput) (*LlmConnectionOutput, error) {
	id, provider, baseURL, apiKey := llmConnectionFromUpdateInput(input)
	c, err := e.uc.UpdateLlmConn.Execute(ctx, id, provider, baseURL, apiKey)
	if err != nil {
		return nil, mapErr(err)
	}
	return toLlmConnectionOutput(c), nil
}

func (e *Endpoints) deleteLlmConnection(ctx context.Context, input *LlmConnectionIDInput) (*struct{}, error) {
	if err := e.uc.DeleteLlmConn.Execute(ctx, input.ID); err != nil {
		return nil, mapErr(err)
	}
	return nil, nil
}

func (e *Endpoints) createModel(ctx context.Context, input *CreateModelInput) (*ModelOutput, error) {
	connID, name, thinkingMode := modelFromCreateInput(input)
	m, err := e.uc.CreateModel.Execute(ctx, connID, name, thinkingMode)
	if err != nil {
		return nil, mapErr(err)
	}
	return toModelOutput(m), nil
}

func (e *Endpoints) listModels(ctx context.Context, _ *struct{}) (*ModelsOutput, error) {
	items, err := e.uc.ListModels.Execute(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return toModelsOutput(items), nil
}

func (e *Endpoints) getModel(ctx context.Context, input *ModelIDInput) (*ModelOutput, error) {
	m, err := e.uc.GetModel.Execute(ctx, input.ID)
	if err != nil {
		return nil, mapErr(err)
	}
	return toModelOutput(m), nil
}

func (e *Endpoints) updateModel(ctx context.Context, input *UpdateModelInput) (*ModelOutput, error) {
	id, connID, name, thinkingMode := modelFromUpdateInput(input)
	m, err := e.uc.UpdateModel.Execute(ctx, id, connID, name, thinkingMode)
	if err != nil {
		return nil, mapErr(err)
	}
	return toModelOutput(m), nil
}

func (e *Endpoints) deleteModel(ctx context.Context, input *ModelIDInput) (*struct{}, error) {
	if err := e.uc.DeleteModel.Execute(ctx, input.ID); err != nil {
		return nil, mapErr(err)
	}
	return nil, nil
}

func (e *Endpoints) createConnection(ctx context.Context, input *CreateConnectionInput) (*ConnectionOutput, error) {
	connType, name, description, modelID, config, profileIDs := connectionFromCreateInput(input)
	c, err := e.uc.CreateConnection.Execute(ctx, connType, name, description, modelID, config, profileIDs)
	if err != nil {
		return nil, mapErr(err)
	}
	return toConnectionOutput(c), nil
}

func (e *Endpoints) listConnections(ctx context.Context, _ *struct{}) (*ConnectionsOutput, error) {
	items, err := e.uc.ListConnections.Execute(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return toConnectionsOutput(items), nil
}

func (e *Endpoints) getConnection(ctx context.Context, input *ConnectionIDInput) (*ConnectionOutput, error) {
	c, err := e.uc.GetConnection.Execute(ctx, input.ID)
	if err != nil {
		return nil, mapErr(err)
	}
	return toConnectionOutput(c), nil
}

func (e *Endpoints) updateConnection(ctx context.Context, input *UpdateConnectionInput) (*ConnectionOutput, error) {
	id, connType, name, description, modelID, config, profileIDs := connectionFromUpdateInput(input)
	c, err := e.uc.UpdateConnection.Execute(ctx, id, connType, name, description, modelID, config, profileIDs)
	if err != nil {
		return nil, mapErr(err)
	}
	return toConnectionOutput(c), nil
}

func (e *Endpoints) deleteConnection(ctx context.Context, input *ConnectionIDInput) (*struct{}, error) {
	if err := e.uc.DeleteConnection.Execute(ctx, input.ID); err != nil {
		return nil, mapErr(err)
	}
	return nil, nil
}

func (e *Endpoints) addMemory(ctx context.Context, input *AddMemoryInput) (*ConnectionOutput, error) {
	c, err := e.uc.AddMemory.Execute(ctx, input.ID, input.Body.Content)
	if err != nil {
		return nil, mapErr(err)
	}
	return toConnectionOutput(c), nil
}

func (e *Endpoints) deleteMemory(ctx context.Context, input *DeleteMemoryInput) (*struct{}, error) {
	_, err := e.uc.DeleteMemory.Execute(ctx, input.ID, input.MemoryID)
	if err != nil {
		return nil, mapErr(err)
	}
	return nil, nil
}

func (e *Endpoints) createCronJob(ctx context.Context, input *CreateCronJobInput) (*CronJobOutput, error) {
	name, schedule, prompt, enabled := cronJobFromCreateInput(input)
	j, err := e.uc.CreateCronJob.Execute(ctx, name, schedule, prompt, enabled)
	if err != nil {
		return nil, mapErr(err)
	}
	return toCronJobOutput(j), nil
}

func (e *Endpoints) listCronJobs(ctx context.Context, _ *struct{}) (*CronJobsOutput, error) {
	items, err := e.uc.ListCronJobs.Execute(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return toCronJobsOutput(items), nil
}

func (e *Endpoints) getCronJob(ctx context.Context, input *CronJobIDInput) (*CronJobOutput, error) {
	j, err := e.uc.GetCronJob.Execute(ctx, input.ID)
	if err != nil {
		return nil, mapErr(err)
	}
	return toCronJobOutput(j), nil
}

func (e *Endpoints) updateCronJob(ctx context.Context, input *UpdateCronJobInput) (*CronJobOutput, error) {
	id, name, schedule, prompt, enabled := cronJobFromUpdateInput(input)
	j, err := e.uc.UpdateCronJob.Execute(ctx, id, name, schedule, prompt, enabled)
	if err != nil {
		return nil, mapErr(err)
	}
	return toCronJobOutput(j), nil
}

func (e *Endpoints) deleteCronJob(ctx context.Context, input *CronJobIDInput) (*struct{}, error) {
	if err := e.uc.DeleteCronJob.Execute(ctx, input.ID); err != nil {
		return nil, mapErr(err)
	}
	return nil, nil
}

func (e *Endpoints) createGuardProfile(ctx context.Context, input *CreateGuardProfileInput) (*GuardProfileOutput, error) {
	name, desc, caps, cmds := guardProfileFromCreateInput(input)
	p, err := e.uc.CreateGuardProfile.Execute(ctx, name, desc, caps, cmds)
	if err != nil {
		return nil, mapErr(err)
	}
	return toGuardProfileOutput(p), nil
}

func (e *Endpoints) listGuardProfiles(ctx context.Context, _ *struct{}) (*GuardProfilesOutput, error) {
	items, err := e.uc.ListGuardProfiles.Execute(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return toGuardProfilesOutput(items), nil
}

func (e *Endpoints) updateGuardProfile(ctx context.Context, input *UpdateGuardProfileInput) (*GuardProfileOutput, error) {
	id, name, desc, caps, cmds := guardProfileFromUpdateInput(input)
	p, err := e.uc.UpdateGuardProfile.Execute(ctx, id, name, desc, caps, cmds)
	if err != nil {
		return nil, mapErr(err)
	}
	return toGuardProfileOutput(p), nil
}

func (e *Endpoints) deleteGuardProfile(ctx context.Context, input *GuardProfileIDInput) (*struct{}, error) {
	if err := e.uc.DeleteGuardProfile.Execute(ctx, input.ID); err != nil {
		return nil, mapErr(err)
	}
	return nil, nil
}

func (e *Endpoints) createChannel(ctx context.Context, input *CreateChannelInput) (*ChannelOutput, error) {
	chType, name, token, modelID, allowed := channelFromCreateInput(input)
	c, err := e.uc.CreateChannel.Execute(ctx, chType, name, token, modelID, allowed)
	if err != nil {
		return nil, mapErr(err)
	}
	return toChannelOutput(c), nil
}

func (e *Endpoints) listChannels(ctx context.Context, _ *struct{}) (*ChannelsOutput, error) {
	items, err := e.uc.ListChannels.Execute(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return toChannelsOutput(items), nil
}

func (e *Endpoints) getChannel(ctx context.Context, input *ChannelIDInput) (*ChannelOutput, error) {
	c, err := e.uc.GetChannel.Execute(ctx, input.ID)
	if err != nil {
		return nil, mapErr(err)
	}
	return toChannelOutput(c), nil
}

func (e *Endpoints) updateChannel(ctx context.Context, input *UpdateChannelInput) (*ChannelOutput, error) {
	id, name, token, modelID, allowed := channelFromUpdateInput(input)
	c, err := e.uc.UpdateChannel.Execute(ctx, id, name, token, modelID, allowed)
	if err != nil {
		return nil, mapErr(err)
	}
	return toChannelOutput(c), nil
}

func (e *Endpoints) deleteChannel(ctx context.Context, input *ChannelIDInput) (*struct{}, error) {
	if err := e.uc.DeleteChannel.Execute(ctx, input.ID); err != nil {
		return nil, mapErr(err)
	}
	return nil, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, base.ErrNotFound):
		return huma.NewError(http.StatusNotFound, err.Error())
	case errors.Is(err, base.ErrValidation):
		return huma.NewError(http.StatusUnprocessableEntity, err.Error())
	default:
		return huma.NewError(http.StatusInternalServerError, err.Error())
	}
}
