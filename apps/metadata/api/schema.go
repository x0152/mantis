package api

import (
	"encoding/json"

	"mantis/core/types"
)

type SettingsOutput struct {
	Body types.Settings
}

type LlmConnectionOutput struct {
	Body types.LlmConnection
}

type LlmConnectionsOutput struct {
	Body []types.LlmConnection
}

type LlmConnectionIDInput struct {
	ID string `path:"id"`
}

type CreateLlmConnectionInput struct {
	Body struct {
		ID       string `json:"id" required:"true" minLength:"1"`
		Provider string `json:"provider" required:"true" minLength:"1"`
		BaseURL  string `json:"baseUrl" required:"true" minLength:"1"`
		APIKey   string `json:"apiKey"`
	}
}

type UpdateLlmConnectionInput struct {
	ID   string `path:"id"`
	Body struct {
		Provider string `json:"provider" required:"true" minLength:"1"`
		BaseURL  string `json:"baseUrl" required:"true" minLength:"1"`
		APIKey   string `json:"apiKey"`
	}
}

type UpdateSettingsInput struct {
	Body struct {
		ChatPresetID   string   `json:"chatPresetId"`
		ServerPresetID string   `json:"serverPresetId"`
		MemoryEnabled  bool     `json:"memoryEnabled"`
		UserMemories   []string `json:"userMemories"`
	}
}

type ModelOutput struct {
	Body types.Model
}

type ModelsOutput struct {
	Body []types.Model
}

type ModelIDInput struct {
	ID string `path:"id"`
}

type CreateModelInput struct {
	Body struct {
		ConnectionID string `json:"connectionId" required:"true" minLength:"1"`
		Name         string `json:"name" required:"true" minLength:"1"`
		ThinkingMode string `json:"thinkingMode" enum:",skip,inline"`
	}
}

type UpdateModelInput struct {
	ID   string `path:"id"`
	Body struct {
		ConnectionID string `json:"connectionId" required:"true" minLength:"1"`
		Name         string `json:"name" required:"true" minLength:"1"`
		ThinkingMode string `json:"thinkingMode" enum:",skip,inline"`
	}
}

type PresetOutput struct {
	Body types.Preset
}

type PresetsOutput struct {
	Body []types.Preset
}

type PresetIDInput struct {
	ID string `path:"id"`
}

type CreatePresetInput struct {
	Body struct {
		Name            string   `json:"name" required:"true" minLength:"1"`
		ChatModelID     string   `json:"chatModelId" required:"true" minLength:"1"`
		SummaryModelID  string   `json:"summaryModelId"`
		ImageModelID    string   `json:"imageModelId"`
		FallbackModelID string   `json:"fallbackModelId"`
		Temperature     *float64 `json:"temperature"`
		SystemPrompt    string   `json:"systemPrompt"`
	}
}

type UpdatePresetInput struct {
	ID   string `path:"id"`
	Body struct {
		Name            string   `json:"name" required:"true" minLength:"1"`
		ChatModelID     string   `json:"chatModelId" required:"true" minLength:"1"`
		SummaryModelID  string   `json:"summaryModelId"`
		ImageModelID    string   `json:"imageModelId"`
		FallbackModelID string   `json:"fallbackModelId"`
		Temperature     *float64 `json:"temperature"`
		SystemPrompt    string   `json:"systemPrompt"`
	}
}

type ConnectionOutput struct {
	Body types.Connection
}

type ConnectionsOutput struct {
	Body []types.Connection
}

type ConnectionIDInput struct {
	ID string `path:"id"`
}

type SkillOutput struct {
	Body types.Skill
}

type SkillsOutput struct {
	Body []types.Skill
}

type SkillIDInput struct {
	ID string `path:"id"`
}

type ListSkillsInput struct {
	ConnectionID string `query:"connectionId"`
}

type CreateSkillInput struct {
	Body struct {
		ConnectionID string          `json:"connectionId" required:"true" minLength:"1"`
		Name         string          `json:"name" required:"true" minLength:"1"`
		Description  string          `json:"description"`
		Parameters   json.RawMessage `json:"parameters"`
		Script       string          `json:"script" required:"true" minLength:"1"`
	}
}

type UpdateSkillInput struct {
	ID   string `path:"id"`
	Body struct {
		ConnectionID string          `json:"connectionId" required:"true" minLength:"1"`
		Name         string          `json:"name" required:"true" minLength:"1"`
		Description  string          `json:"description"`
		Parameters   json.RawMessage `json:"parameters"`
		Script       string          `json:"script" required:"true" minLength:"1"`
	}
}

type CreateConnectionInput struct {
	Body struct {
		Type          string          `json:"type" required:"true" enum:"ssh"`
		Name          string          `json:"name" required:"true" minLength:"1"`
		Description   string          `json:"description"`
		ModelID       string          `json:"modelId"`
		PresetID      string          `json:"presetId"`
		Config        json.RawMessage `json:"config" required:"true"`
		ProfileIDs    []string        `json:"profileIds"`
		MemoryEnabled *bool           `json:"memoryEnabled"`
	}
}

type UpdateConnectionInput struct {
	ID   string `path:"id"`
	Body struct {
		Type          string          `json:"type" required:"true" enum:"ssh"`
		Name          string          `json:"name" required:"true" minLength:"1"`
		Description   string          `json:"description"`
		ModelID       string          `json:"modelId"`
		PresetID      string          `json:"presetId"`
		Config        json.RawMessage `json:"config" required:"true"`
		ProfileIDs    []string        `json:"profileIds"`
		MemoryEnabled *bool           `json:"memoryEnabled"`
	}
}

type AddMemoryInput struct {
	ID   string `path:"id"`
	Body struct {
		Content string `json:"content" required:"true" minLength:"1"`
	}
}

type DeleteMemoryInput struct {
	ID       string `path:"id"`
	MemoryID string `path:"memoryId"`
}

type CronJobOutput struct {
	Body types.CronJob
}

type CronJobsOutput struct {
	Body []types.CronJob
}

type CronJobIDInput struct {
	ID string `path:"id"`
}

type CreateCronJobInput struct {
	Body struct {
		Name     string `json:"name" required:"true" minLength:"1"`
		Schedule string `json:"schedule" required:"true" minLength:"1"`
		Prompt   string `json:"prompt" required:"true"`
		Enabled  bool   `json:"enabled"`
	}
}

type UpdateCronJobInput struct {
	ID   string `path:"id"`
	Body struct {
		Name     string `json:"name" required:"true" minLength:"1"`
		Schedule string `json:"schedule" required:"true" minLength:"1"`
		Prompt   string `json:"prompt" required:"true"`
		Enabled  bool   `json:"enabled"`
	}
}

type GuardProfileOutput struct {
	Body types.GuardProfile
}

type GuardProfilesOutput struct {
	Body []types.GuardProfile
}

type GuardProfileIDInput struct {
	ID string `path:"id"`
}

type CreateGuardProfileInput struct {
	Body struct {
		Name         string                  `json:"name" required:"true" minLength:"1"`
		Description  string                  `json:"description"`
		Capabilities types.GuardCapabilities `json:"capabilities"`
		Commands     []types.CommandRule     `json:"commands"`
	}
}

type UpdateGuardProfileInput struct {
	ID   string `path:"id"`
	Body struct {
		Name         string                  `json:"name" required:"true" minLength:"1"`
		Description  string                  `json:"description"`
		Capabilities types.GuardCapabilities `json:"capabilities"`
		Commands     []types.CommandRule     `json:"commands"`
	}
}

type ChannelOutput struct {
	Body types.Channel
}

type ChannelsOutput struct {
	Body []types.Channel
}

type ChannelIDInput struct {
	ID string `path:"id"`
}

type CreateChannelInput struct {
	Body struct {
		Type           string  `json:"type" required:"true" enum:"telegram"`
		Name           string  `json:"name"`
		Token          string  `json:"token" required:"true" minLength:"1"`
		ModelID        string  `json:"modelId"`
		PresetID       string  `json:"presetId"`
		AllowedUserIDs []int64 `json:"allowedUserIds"`
	}
}

type UpdateChannelInput struct {
	ID   string `path:"id"`
	Body struct {
		Name           string  `json:"name"`
		Token          string  `json:"token"`
		ModelID        string  `json:"modelId"`
		PresetID       string  `json:"presetId"`
		AllowedUserIDs []int64 `json:"allowedUserIds"`
	}
}
