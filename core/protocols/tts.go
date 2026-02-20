package protocols

import "context"

type TTSRequest struct {
	Text         string
	Voice        []byte
	Emotion      string
	Instructions string
	Format       string
}

type TTS interface {
	Synthesize(ctx context.Context, req TTSRequest) ([]byte, error)
}
