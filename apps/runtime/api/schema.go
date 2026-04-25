package api

import "mantis/core/types"

type RunInput struct {
	Name    string            `json:"name"`
	Image   string            `json:"image,omitempty"`
	Network string            `json:"network,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
	Cmd     []string          `json:"cmd,omitempty"`
}

type ContainerOutput struct {
	Container types.RuntimeContainer `json:"container"`
}

type ContainerListOutput struct {
	Containers []types.RuntimeContainer `json:"containers"`
}

type RegisterInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	ProfileIDs  []string `json:"profileIds,omitempty"`
	Username    string   `json:"username,omitempty"`
	Password    string   `json:"password,omitempty"`
	Port        int      `json:"port,omitempty"`
}

type RegisterOutput struct {
	Connection types.Connection `json:"connection"`
}
