package types

type ProviderModel struct {
	ID string `json:"id"`
}

type InferenceLimit struct {
	Type       string `json:"type"`
	Percentage *int   `json:"percentage,omitempty"`
	Label      string `json:"label"`
}
