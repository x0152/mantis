package types

import "time"

type RuntimeContainer struct {
	Name      string            `json:"name"`
	Image     string            `json:"image"`
	Status    string            `json:"status"`
	Host      string            `json:"host"`
	Port      int               `json:"port,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt time.Time         `json:"createdAt"`
}

type RuntimeRunSpec struct {
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Network string            `json:"network,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Cmd     []string          `json:"cmd,omitempty"`
}
