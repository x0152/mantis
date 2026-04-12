package protocols

import (
	"context"

	"mantis/core/types"
)

type PlanRunner interface {
	TriggerRun(ctx context.Context, planID, trigger string) (types.PlanRun, error)
}
