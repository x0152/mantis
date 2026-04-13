package protocols

import (
	"context"

	"mantis/core/types"
)

type PlanRunner interface {
	TriggerRun(ctx context.Context, planID, trigger string, input map[string]any) (types.PlanRun, error)
	CancelRun(ctx context.Context, runID string) (types.PlanRun, error)
	ActiveRuns(ctx context.Context) ([]types.PlanRun, error)
}
