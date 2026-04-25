package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "mantis/apps/metadata/use_cases"
	"mantis/apps/plans"
	"mantis/core/base"
	"mantis/core/types"
)

type UseCases struct {
	GetSettings        *usecases.GetSettings
	UpdateSettings     *usecases.UpdateSettings
	CreateLlmConn      *usecases.CreateLlmConnection
	GetLlmConn         *usecases.GetLlmConnection
	ListLlmConns       *usecases.ListLlmConnections
	UpdateLlmConn      *usecases.UpdateLlmConnection
	DeleteLlmConn      *usecases.DeleteLlmConnection
	ListConnModels     *usecases.ListConnectionModels
	GetConnLimit       *usecases.GetConnectionInferenceLimit
	CreateModel        *usecases.CreateModel
	GetModel           *usecases.GetModel
	ListModels         *usecases.ListModels
	UpdateModel        *usecases.UpdateModel
	DeleteModel        *usecases.DeleteModel
	CreatePreset       *usecases.CreatePreset
	ListPresets        *usecases.ListPresets
	UpdatePreset       *usecases.UpdatePreset
	DeletePreset       *usecases.DeletePreset
	CreateConnection   *usecases.CreateConnection
	GetConnection      *usecases.GetConnection
	ListConnections    *usecases.ListConnections
	UpdateConnection   *usecases.UpdateConnection
	DeleteConnection   *usecases.DeleteConnection
	CreateSkill        *usecases.CreateSkill
	ListSkills         *usecases.ListSkills
	UpdateSkill        *usecases.UpdateSkill
	DeleteSkill        *usecases.DeleteSkill
	AddMemory          *usecases.AddMemory
	DeleteMemory       *usecases.DeleteMemory
	CreatePlan         *usecases.CreatePlan
	ListPlans          *usecases.ListPlans
	UpdatePlan         *usecases.UpdatePlan
	DeletePlan         *usecases.DeletePlan
	ListPlanRuns       *usecases.ListPlanRuns
	GetPlanRun         *usecases.GetPlanRun
	PlanRunner         *plans.Runner
	CreateGuardProfile *usecases.CreateGuardProfile
	ListGuardProfiles  *usecases.ListGuardProfiles
	UpdateGuardProfile *usecases.UpdateGuardProfile
	DeleteGuardProfile *usecases.DeleteGuardProfile
	CreateChannel      *usecases.CreateChannel
	GetChannel         *usecases.GetChannel
	ListChannels       *usecases.ListChannels
	UpdateChannel      *usecases.UpdateChannel
	DeleteChannel      *usecases.DeleteChannel
}

type Endpoints struct {
	uc UseCases
}

func NewEndpoints(uc UseCases) *Endpoints {
	return &Endpoints{uc: uc}
}

func (e *Endpoints) Register(api huma.API) {
	huma.Register(api, huma.Operation{OperationID: "get-settings", Method: http.MethodGet, Path: "/api/settings"}, e.getSettings)
	huma.Register(api, huma.Operation{OperationID: "update-settings", Method: http.MethodPut, Path: "/api/settings"}, e.updateSettings)

	huma.Register(api, huma.Operation{OperationID: "create-llm-connection", Method: http.MethodPost, Path: "/api/llm-connections", DefaultStatus: 201}, e.createLlmConnection)
	huma.Register(api, huma.Operation{OperationID: "list-llm-connections", Method: http.MethodGet, Path: "/api/llm-connections"}, e.listLlmConnections)
	huma.Register(api, huma.Operation{OperationID: "get-llm-connection", Method: http.MethodGet, Path: "/api/llm-connections/{id}"}, e.getLlmConnection)
	huma.Register(api, huma.Operation{OperationID: "update-llm-connection", Method: http.MethodPut, Path: "/api/llm-connections/{id}"}, e.updateLlmConnection)
	huma.Register(api, huma.Operation{OperationID: "delete-llm-connection", Method: http.MethodDelete, Path: "/api/llm-connections/{id}", DefaultStatus: 204}, e.deleteLlmConnection)
	huma.Register(api, huma.Operation{OperationID: "list-llm-connection-models", Method: http.MethodGet, Path: "/api/llm-connections/{id}/available-models"}, e.listLlmConnectionModels)
	huma.Register(api, huma.Operation{OperationID: "get-llm-connection-inference-limit", Method: http.MethodGet, Path: "/api/llm-connections/{id}/inference-limit"}, e.getLlmConnectionInferenceLimit)

	huma.Register(api, huma.Operation{OperationID: "create-model", Method: http.MethodPost, Path: "/api/models", DefaultStatus: 201}, e.createModel)
	huma.Register(api, huma.Operation{OperationID: "list-models", Method: http.MethodGet, Path: "/api/models"}, e.listModels)
	huma.Register(api, huma.Operation{OperationID: "get-model", Method: http.MethodGet, Path: "/api/models/{id}"}, e.getModel)
	huma.Register(api, huma.Operation{OperationID: "update-model", Method: http.MethodPut, Path: "/api/models/{id}"}, e.updateModel)
	huma.Register(api, huma.Operation{OperationID: "delete-model", Method: http.MethodDelete, Path: "/api/models/{id}", DefaultStatus: 204}, e.deleteModel)

	huma.Register(api, huma.Operation{OperationID: "create-preset", Method: http.MethodPost, Path: "/api/presets", DefaultStatus: 201}, e.createPreset)
	huma.Register(api, huma.Operation{OperationID: "list-presets", Method: http.MethodGet, Path: "/api/presets"}, e.listPresets)
	huma.Register(api, huma.Operation{OperationID: "update-preset", Method: http.MethodPut, Path: "/api/presets/{id}"}, e.updatePreset)
	huma.Register(api, huma.Operation{OperationID: "delete-preset", Method: http.MethodDelete, Path: "/api/presets/{id}", DefaultStatus: 204}, e.deletePreset)

	huma.Register(api, huma.Operation{OperationID: "create-connection", Method: http.MethodPost, Path: "/api/connections", DefaultStatus: 201}, e.createConnection)
	huma.Register(api, huma.Operation{OperationID: "list-connections", Method: http.MethodGet, Path: "/api/connections"}, e.listConnections)
	huma.Register(api, huma.Operation{OperationID: "get-connection", Method: http.MethodGet, Path: "/api/connections/{id}"}, e.getConnection)
	huma.Register(api, huma.Operation{OperationID: "update-connection", Method: http.MethodPut, Path: "/api/connections/{id}"}, e.updateConnection)
	huma.Register(api, huma.Operation{OperationID: "delete-connection", Method: http.MethodDelete, Path: "/api/connections/{id}", DefaultStatus: 204}, e.deleteConnection)
	huma.Register(api, huma.Operation{OperationID: "create-skill", Method: http.MethodPost, Path: "/api/skills", DefaultStatus: 201}, e.createSkill)
	huma.Register(api, huma.Operation{OperationID: "list-skills", Method: http.MethodGet, Path: "/api/skills"}, e.listSkills)
	huma.Register(api, huma.Operation{OperationID: "update-skill", Method: http.MethodPut, Path: "/api/skills/{id}"}, e.updateSkill)
	huma.Register(api, huma.Operation{OperationID: "delete-skill", Method: http.MethodDelete, Path: "/api/skills/{id}", DefaultStatus: 204}, e.deleteSkill)
	huma.Register(api, huma.Operation{OperationID: "add-memory", Method: http.MethodPost, Path: "/api/connections/{id}/memories", DefaultStatus: 201}, e.addMemory)
	huma.Register(api, huma.Operation{OperationID: "delete-memory", Method: http.MethodDelete, Path: "/api/connections/{id}/memories/{memoryId}", DefaultStatus: 204}, e.deleteMemory)

	huma.Register(api, huma.Operation{OperationID: "create-plan", Method: http.MethodPost, Path: "/api/plans", DefaultStatus: 201}, e.createPlan)
	huma.Register(api, huma.Operation{OperationID: "list-plans", Method: http.MethodGet, Path: "/api/plans"}, e.listPlans)
	huma.Register(api, huma.Operation{OperationID: "update-plan", Method: http.MethodPut, Path: "/api/plans/{id}"}, e.updatePlan)
	huma.Register(api, huma.Operation{OperationID: "delete-plan", Method: http.MethodDelete, Path: "/api/plans/{id}", DefaultStatus: 204}, e.deletePlan)

	huma.Register(api, huma.Operation{OperationID: "list-plan-runs", Method: http.MethodGet, Path: "/api/plans/{planId}/runs"}, e.listPlanRuns)
	huma.Register(api, huma.Operation{OperationID: "trigger-plan-run", Method: http.MethodPost, Path: "/api/plans/{planId}/runs", DefaultStatus: 201}, e.triggerPlanRun)
	huma.Register(api, huma.Operation{OperationID: "get-plan-run", Method: http.MethodGet, Path: "/api/plan-runs/{id}"}, e.getPlanRun)
	huma.Register(api, huma.Operation{OperationID: "cancel-plan-run", Method: http.MethodPost, Path: "/api/plan-runs/{id}/cancel"}, e.cancelPlanRun)

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

func (e *Endpoints) getSettings(ctx context.Context, _ *struct{}) (*SettingsOutput, error) {
	s, err := e.uc.GetSettings.Execute(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return toSettingsOutput(s), nil
}

func (e *Endpoints) updateSettings(ctx context.Context, input *UpdateSettingsInput) (*SettingsOutput, error) {
	s, err := e.uc.UpdateSettings.Execute(ctx, settingsFromInput(input))
	if err != nil {
		return nil, mapErr(err)
	}
	return toSettingsOutput(s), nil
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

func (e *Endpoints) listLlmConnectionModels(ctx context.Context, input *LlmConnectionIDInput) (*ProviderModelsOutput, error) {
	items, err := e.uc.ListConnModels.Execute(ctx, input.ID)
	if err != nil {
		return nil, mapErr(err)
	}
	return toProviderModelsOutput(items), nil
}

func (e *Endpoints) getLlmConnectionInferenceLimit(ctx context.Context, input *LlmConnectionIDInput) (*InferenceLimitOutput, error) {
	limit, err := e.uc.GetConnLimit.Execute(ctx, input.ID)
	if err != nil {
		return nil, mapErr(err)
	}
	return toInferenceLimitOutput(limit), nil
}

func (e *Endpoints) createModel(ctx context.Context, input *CreateModelInput) (*ModelOutput, error) {
	m, err := e.uc.CreateModel.Execute(ctx, modelFromCreateInput(input))
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
	m, err := e.uc.UpdateModel.Execute(ctx, modelFromUpdateInput(input))
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

func (e *Endpoints) createPreset(ctx context.Context, input *CreatePresetInput) (*PresetOutput, error) {
	p, err := e.uc.CreatePreset.Execute(ctx, types.Preset{
		Name:            input.Body.Name,
		ChatModelID:     input.Body.ChatModelID,
		SummaryModelID:  input.Body.SummaryModelID,
		ImageModelID:    input.Body.ImageModelID,
		FallbackModelID: input.Body.FallbackModelID,
		Temperature:     input.Body.Temperature,
		SystemPrompt:    input.Body.SystemPrompt,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return toPresetOutput(p), nil
}

func (e *Endpoints) listPresets(ctx context.Context, _ *struct{}) (*PresetsOutput, error) {
	items, err := e.uc.ListPresets.Execute(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return toPresetsOutput(items), nil
}

func (e *Endpoints) updatePreset(ctx context.Context, input *UpdatePresetInput) (*PresetOutput, error) {
	p, err := e.uc.UpdatePreset.Execute(ctx, types.Preset{
		ID:              input.ID,
		Name:            input.Body.Name,
		ChatModelID:     input.Body.ChatModelID,
		SummaryModelID:  input.Body.SummaryModelID,
		ImageModelID:    input.Body.ImageModelID,
		FallbackModelID: input.Body.FallbackModelID,
		Temperature:     input.Body.Temperature,
		SystemPrompt:    input.Body.SystemPrompt,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return toPresetOutput(p), nil
}

func (e *Endpoints) deletePreset(ctx context.Context, input *PresetIDInput) (*struct{}, error) {
	if err := e.uc.DeletePreset.Execute(ctx, input.ID); err != nil {
		return nil, mapErr(err)
	}
	return nil, nil
}

func (e *Endpoints) createConnection(ctx context.Context, input *CreateConnectionInput) (*ConnectionOutput, error) {
	connType, name, description, modelID, presetID, config, profileIDs, memoryEnabled := connectionFromCreateInput(input)
	c, err := e.uc.CreateConnection.Execute(ctx, connType, name, description, modelID, presetID, config, profileIDs, memoryEnabled)
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
	id, connType, name, description, modelID, presetID, config, profileIDs, memoryEnabled := connectionFromUpdateInput(input)
	c, err := e.uc.UpdateConnection.Execute(ctx, id, connType, name, description, modelID, presetID, config, profileIDs, memoryEnabled)
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

func (e *Endpoints) createSkill(ctx context.Context, input *CreateSkillInput) (*SkillOutput, error) {
	s, err := e.uc.CreateSkill.Execute(ctx, skillFromCreateInput(input))
	if err != nil {
		return nil, mapErr(err)
	}
	return toSkillOutput(s), nil
}

func (e *Endpoints) listSkills(ctx context.Context, input *ListSkillsInput) (*SkillsOutput, error) {
	items, err := e.uc.ListSkills.Execute(ctx, input.ConnectionID)
	if err != nil {
		return nil, mapErr(err)
	}
	return toSkillsOutput(items), nil
}

func (e *Endpoints) updateSkill(ctx context.Context, input *UpdateSkillInput) (*SkillOutput, error) {
	s, err := e.uc.UpdateSkill.Execute(ctx, skillFromUpdateInput(input))
	if err != nil {
		return nil, mapErr(err)
	}
	return toSkillOutput(s), nil
}

func (e *Endpoints) deleteSkill(ctx context.Context, input *SkillIDInput) (*struct{}, error) {
	if err := e.uc.DeleteSkill.Execute(ctx, input.ID); err != nil {
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

func (e *Endpoints) createPlan(ctx context.Context, input *CreatePlanInput) (*PlanOutput, error) {
	p, err := e.uc.CreatePlan.Execute(ctx, planFromCreateInput(input))
	if err != nil {
		return nil, mapErr(err)
	}
	return toPlanOutput(p), nil
}

func (e *Endpoints) listPlans(ctx context.Context, _ *struct{}) (*PlansOutput, error) {
	items, err := e.uc.ListPlans.Execute(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return toPlansOutput(items), nil
}

func (e *Endpoints) updatePlan(ctx context.Context, input *UpdatePlanInput) (*PlanOutput, error) {
	p, err := e.uc.UpdatePlan.Execute(ctx, planFromUpdateInput(input))
	if err != nil {
		return nil, mapErr(err)
	}
	return toPlanOutput(p), nil
}

func (e *Endpoints) deletePlan(ctx context.Context, input *PlanIDInput) (*struct{}, error) {
	if err := e.uc.DeletePlan.Execute(ctx, input.ID); err != nil {
		return nil, mapErr(err)
	}
	return nil, nil
}

func (e *Endpoints) listPlanRuns(ctx context.Context, input *ListPlanRunsInput) (*PlanRunsOutput, error) {
	items, err := e.uc.ListPlanRuns.Execute(ctx, input.PlanID)
	if err != nil {
		return nil, mapErr(err)
	}
	return toPlanRunsOutput(items), nil
}

func (e *Endpoints) triggerPlanRun(ctx context.Context, input *TriggerPlanRunInput) (*PlanRunOutput, error) {
	if e.uc.PlanRunner == nil {
		return nil, huma.NewError(http.StatusServiceUnavailable, "plan runner not available")
	}
	run, err := e.uc.PlanRunner.TriggerRun(ctx, input.PlanID, "manual", input.Body.Input)
	if err != nil {
		return nil, mapErr(err)
	}
	return toPlanRunOutput(run), nil
}

func (e *Endpoints) getPlanRun(ctx context.Context, input *PlanRunIDInput) (*PlanRunOutput, error) {
	r, err := e.uc.GetPlanRun.Execute(ctx, input.ID)
	if err != nil {
		return nil, mapErr(err)
	}
	return toPlanRunOutput(r), nil
}

func (e *Endpoints) cancelPlanRun(ctx context.Context, input *CancelPlanRunInput) (*PlanRunOutput, error) {
	if e.uc.PlanRunner == nil {
		return nil, huma.NewError(http.StatusServiceUnavailable, "plan runner not available")
	}
	run, err := e.uc.PlanRunner.CancelRun(ctx, input.ID)
	if err != nil {
		return nil, mapErr(err)
	}
	return toPlanRunOutput(run), nil
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
	chType, name, token, modelID, presetID, allowed := channelFromCreateInput(input)
	c, err := e.uc.CreateChannel.Execute(ctx, chType, name, token, modelID, presetID, allowed)
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
	id, name, token, modelID, presetID, allowed := channelFromUpdateInput(input)
	c, err := e.uc.UpdateChannel.Execute(ctx, id, name, token, modelID, presetID, allowed)
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
