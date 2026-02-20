package types

type SortDir string

const (
	SortDirAsc  SortDir = "asc"
	SortDirDesc SortDir = "desc"
)

type Sort struct {
	Field string  `json:"field"`
	Dir   SortDir `json:"dir"`
}

type ListQuery struct {
	Page      Page              `json:"page,omitempty"`
	Filter    map[string]string `json:"filter,omitempty"`
	FilterNot map[string]string `json:"filterNot,omitempty"`
	Sort      []Sort            `json:"sort,omitempty"`
}
