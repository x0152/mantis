package types

type GuardCapabilities struct {
	Pipes       bool `json:"pipes"`
	Redirects   bool `json:"redirects"`
	CmdSubst    bool `json:"cmdSubst"`
	Background  bool `json:"background"`
	Sudo        bool `json:"sudo"`
	CodeExec    bool `json:"codeExec"`
	Download    bool `json:"download"`
	Install     bool `json:"install"`
	WriteFS     bool `json:"writeFs"`
	NetworkOut  bool `json:"networkOut"`
	Cron        bool `json:"cron"`
	Unrestricted bool `json:"unrestricted"`
}

type CommandRule struct {
	Command    string   `json:"command"`
	AllowedArgs []string `json:"allowedArgs,omitempty"`
	AllowedSQL  []string `json:"allowedSql,omitempty"`
}

type GuardProfile struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Builtin      bool              `json:"builtin"`
	Capabilities GuardCapabilities `json:"capabilities"`
	Commands     []CommandRule     `json:"commands"`
}
