package protocols

import (
	"context"
	"io"

	"mantis/core/types"
)

type Runtime interface {
	Build(ctx context.Context, name string, dockerfile []byte) (io.ReadCloser, error)
	Run(ctx context.Context, spec types.RuntimeRunSpec) (types.RuntimeContainer, error)
	Stop(ctx context.Context, name string) error
	Remove(ctx context.Context, name string) error
	List(ctx context.Context) ([]types.RuntimeContainer, error)
	Inspect(ctx context.Context, name string) (types.RuntimeContainer, error)
	Logs(ctx context.Context, name string, tail int, follow bool) (io.ReadCloser, error)
}
