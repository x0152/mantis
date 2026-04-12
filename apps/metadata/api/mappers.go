package api

import (
	"encoding/json"

	"mantis/core/types"
)

func toPresetOutput(p types.Preset) *PresetOutput {
	return &PresetOutput{Body: p}
}

func toPresetsOutput(items []types.Preset) *PresetsOutput {
	return &PresetsOutput{Body: items}
}

func toSettingsOutput(s types.Settings) *SettingsOutput {
	return &SettingsOutput{Body: s}
}

func toLlmConnectionOutput(c types.LlmConnection) *LlmConnectionOutput {
	return &LlmConnectionOutput{Body: c}
}

func toLlmConnectionsOutput(items []types.LlmConnection) *LlmConnectionsOutput {
	return &LlmConnectionsOutput{Body: items}
}

func toModelOutput(m types.Model) *ModelOutput {
	return &ModelOutput{Body: m}
}

func toModelsOutput(items []types.Model) *ModelsOutput {
	return &ModelsOutput{Body: items}
}

func toConnectionOutput(c types.Connection) *ConnectionOutput {
	return &ConnectionOutput{Body: c}
}

func toConnectionsOutput(items []types.Connection) *ConnectionsOutput {
	return &ConnectionsOutput{Body: items}
}

func toSkillOutput(s types.Skill) *SkillOutput {
	return &SkillOutput{Body: s}
}

func toSkillsOutput(items []types.Skill) *SkillsOutput {
	return &SkillsOutput{Body: items}
}

func settingsFromInput(input *UpdateSettingsInput) types.Settings {
	return types.Settings{
		ChatPresetID:   input.Body.ChatPresetID,
		ServerPresetID: input.Body.ServerPresetID,
		MemoryEnabled:  input.Body.MemoryEnabled,
		UserMemories:   input.Body.UserMemories,
	}
}

func llmConnectionFromCreateInput(input *CreateLlmConnectionInput) (string, string, string, string) {
	return input.Body.ID, input.Body.Provider, input.Body.BaseURL, input.Body.APIKey
}

func llmConnectionFromUpdateInput(input *UpdateLlmConnectionInput) (string, string, string, string) {
	return input.ID, input.Body.Provider, input.Body.BaseURL, input.Body.APIKey
}

func modelFromCreateInput(input *CreateModelInput) (string, string, string) {
	return input.Body.ConnectionID, input.Body.Name, input.Body.ThinkingMode
}

func modelFromUpdateInput(input *UpdateModelInput) (string, string, string, string) {
	return input.ID, input.Body.ConnectionID, input.Body.Name, input.Body.ThinkingMode
}

func connectionFromCreateInput(input *CreateConnectionInput) (string, string, string, string, string, json.RawMessage, []string, bool) {
	memoryEnabled := true
	if input.Body.MemoryEnabled != nil {
		memoryEnabled = *input.Body.MemoryEnabled
	}
	return input.Body.Type, input.Body.Name, input.Body.Description, input.Body.ModelID, input.Body.PresetID, input.Body.Config, input.Body.ProfileIDs, memoryEnabled
}

func connectionFromUpdateInput(input *UpdateConnectionInput) (string, string, string, string, string, string, json.RawMessage, []string, bool) {
	memoryEnabled := true
	if input.Body.MemoryEnabled != nil {
		memoryEnabled = *input.Body.MemoryEnabled
	}
	return input.ID, input.Body.Type, input.Body.Name, input.Body.Description, input.Body.ModelID, input.Body.PresetID, input.Body.Config, input.Body.ProfileIDs, memoryEnabled
}

func skillFromCreateInput(input *CreateSkillInput) types.Skill {
	return types.Skill{
		ConnectionID: input.Body.ConnectionID,
		Name:         input.Body.Name,
		Description:  input.Body.Description,
		Parameters:   input.Body.Parameters,
		Script:       input.Body.Script,
	}
}

func skillFromUpdateInput(input *UpdateSkillInput) types.Skill {
	return types.Skill{
		ID:           input.ID,
		ConnectionID: input.Body.ConnectionID,
		Name:         input.Body.Name,
		Description:  input.Body.Description,
		Parameters:   input.Body.Parameters,
		Script:       input.Body.Script,
	}
}

func toPlanOutput(p types.Plan) *PlanOutput {
	return &PlanOutput{Body: p}
}

func toPlansOutput(items []types.Plan) *PlansOutput {
	return &PlansOutput{Body: items}
}

func planFromCreateInput(input *CreatePlanInput) types.Plan {
	return types.Plan{
		Name:        input.Body.Name,
		Description: input.Body.Description,
		Schedule:    input.Body.Schedule,
		Enabled:     input.Body.Enabled,
		Graph:       input.Body.Graph,
	}
}

func planFromUpdateInput(input *UpdatePlanInput) types.Plan {
	return types.Plan{
		ID:          input.ID,
		Name:        input.Body.Name,
		Description: input.Body.Description,
		Schedule:    input.Body.Schedule,
		Enabled:     input.Body.Enabled,
		Graph:       input.Body.Graph,
	}
}

func toPlanRunOutput(r types.PlanRun) *PlanRunOutput {
	return &PlanRunOutput{Body: r}
}

func toPlanRunsOutput(items []types.PlanRun) *PlanRunsOutput {
	return &PlanRunsOutput{Body: items}
}

func toCronJobOutput(j types.CronJob) *CronJobOutput {
	return &CronJobOutput{Body: j}
}

func toCronJobsOutput(items []types.CronJob) *CronJobsOutput {
	return &CronJobsOutput{Body: items}
}

func cronJobFromCreateInput(input *CreateCronJobInput) (string, string, string, bool) {
	return input.Body.Name, input.Body.Schedule, input.Body.Prompt, input.Body.Enabled
}

func cronJobFromUpdateInput(input *UpdateCronJobInput) (string, string, string, string, bool) {
	return input.ID, input.Body.Name, input.Body.Schedule, input.Body.Prompt, input.Body.Enabled
}

func toGuardProfileOutput(p types.GuardProfile) *GuardProfileOutput {
	return &GuardProfileOutput{Body: p}
}

func toGuardProfilesOutput(items []types.GuardProfile) *GuardProfilesOutput {
	return &GuardProfilesOutput{Body: items}
}

func guardProfileFromCreateInput(input *CreateGuardProfileInput) (string, string, types.GuardCapabilities, []types.CommandRule) {
	return input.Body.Name, input.Body.Description, input.Body.Capabilities, input.Body.Commands
}

func guardProfileFromUpdateInput(input *UpdateGuardProfileInput) (string, string, string, types.GuardCapabilities, []types.CommandRule) {
	return input.ID, input.Body.Name, input.Body.Description, input.Body.Capabilities, input.Body.Commands
}

func toChannelOutput(c types.Channel) *ChannelOutput {
	return &ChannelOutput{Body: c}
}

func toChannelsOutput(items []types.Channel) *ChannelsOutput {
	return &ChannelsOutput{Body: items}
}

func channelFromCreateInput(input *CreateChannelInput) (string, string, string, string, string, []int64) {
	return input.Body.Type, input.Body.Name, input.Body.Token, input.Body.ModelID, input.Body.PresetID, input.Body.AllowedUserIDs
}

func channelFromUpdateInput(input *UpdateChannelInput) (string, string, string, string, string, []int64) {
	return input.ID, input.Body.Name, input.Body.Token, input.Body.ModelID, input.Body.PresetID, input.Body.AllowedUserIDs
}
