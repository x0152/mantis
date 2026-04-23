package usecases

import (
	"context"
	"strings"

	"mantis/core/plugins/pipeline"
	"mantis/core/protocols"
)

type StopGeneration struct {
	cancellations *pipeline.Cancellations
	planRunner    protocols.PlanRunner
}

func NewStopGeneration(cancellations *pipeline.Cancellations, planRunner protocols.PlanRunner) *StopGeneration {
	return &StopGeneration{cancellations: cancellations, planRunner: planRunner}
}

func (uc *StopGeneration) Execute(ctx context.Context, sessionID string) (bool, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return false, nil
	}

	if strings.HasPrefix(sessionID, "plan:") {
		parts := strings.Split(sessionID, ":")
		if len(parts) >= 3 && uc.planRunner != nil {
			runID := parts[2]
			if _, err := uc.planRunner.CancelRun(ctx, runID); err != nil {
				return false, err
			}
			return true, nil
		}
	}

	return uc.cancellations.Cancel(sessionID), nil
}
