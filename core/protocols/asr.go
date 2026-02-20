package protocols

import (
	"context"
	"io"
)

type ASR interface {
	Transcribe(ctx context.Context, audio io.Reader, format string) (string, error)
}
