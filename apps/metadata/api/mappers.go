package api

import (
	"encoding/json"

	"mantis/core/types"
)

func toConfigOutput(cfg types.Config) *ConfigOutput {
	return &ConfigOutput{Body: cfg}
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

func configDataFromInput(input *UpdateConfigInput) json.RawMessage {
	return input.Body.Data
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

func connectionFromCreateInput(input *CreateConnectionInput) (string, string, string, string, json.RawMessage, []string) {
	return input.Body.Type, input.Body.Name, input.Body.Description, input.Body.ModelID, input.Body.Config, input.Body.ProfileIDs
}

func connectionFromUpdateInput(input *UpdateConnectionInput) (string, string, string, string, string, json.RawMessage, []string) {
	return input.ID, input.Body.Type, input.Body.Name, input.Body.Description, input.Body.ModelID, input.Body.Config, input.Body.ProfileIDs
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

func channelFromCreateInput(input *CreateChannelInput) (string, string, string, string, []int64) {
	return input.Body.Type, input.Body.Name, input.Body.Token, input.Body.ModelID, input.Body.AllowedUserIDs
}

func channelFromUpdateInput(input *UpdateChannelInput) (string, string, string, string, []int64) {
	return input.ID, input.Body.Name, input.Body.Token, input.Body.ModelID, input.Body.AllowedUserIDs
}
