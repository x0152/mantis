package protocols

import "context"

type VisionLLM interface {
	Describe(ctx context.Context, baseURL, apiKey, model string, image []byte, format, prompt string) (string, error)
}
