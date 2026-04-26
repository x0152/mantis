package api

import "mantis/core/types"

type SandboxInput struct {
	Name           string   `json:"name"`
	ConnectionName string   `json:"connectionName,omitempty"`
	Description    string   `json:"description,omitempty"`
	ProfileIDs     []string `json:"profileIds,omitempty"`
	Dockerfile     string   `json:"dockerfile"`
}

type SandboxStatus struct {
	Connection types.Connection       `json:"connection"`
	Container  types.RuntimeContainer `json:"container"`
	State      string                 `json:"state"`
}

type SandboxListOutput struct {
	Sandboxes []SandboxStatus `json:"sandboxes"`
}
