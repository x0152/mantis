package protocols

import (
	"context"
	"io"
)

type OCR interface {
	ExtractText(ctx context.Context, image io.Reader, format string) (string, error)
}
