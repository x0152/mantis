package protocols

import "context"

type Channel interface {
	Execute(ctx context.Context) error
}
