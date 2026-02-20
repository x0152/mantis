package protocols

import (
	"context"

	"mantis/core/types"
)

type Store[ID comparable, Entity any] interface {
	Create(ctx context.Context, items []Entity) ([]Entity, error)
	Get(ctx context.Context, ids []ID) (map[ID]Entity, error)
	List(ctx context.Context, query types.ListQuery) ([]Entity, error)
	Update(ctx context.Context, items []Entity) ([]Entity, error)
	Delete(ctx context.Context, ids []ID) error
}
