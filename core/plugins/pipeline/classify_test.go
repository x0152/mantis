package pipeline

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"mantis/shared"
)

func newTestPipeline() *RequestHandlePipeline {
	return &RequestHandlePipeline{limits: shared.Limits{
		SupervisorMaxIterations: 7,
		SupervisorTimeout:       2 * time.Minute,
	}}
}

func TestClassifyStop_UserCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	marker, stopped := newTestPipeline().classifyStop(ctx, nil)
	if !stopped || !strings.Contains(marker, "stopped by user") {
		t.Fatalf("expected user marker, got stopped=%v marker=%q", stopped, marker)
	}
}

func TestClassifyStop_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond)
	defer cancel()
	time.Sleep(2 * time.Millisecond)
	marker, stopped := newTestPipeline().classifyStop(ctx, nil)
	if !stopped || !strings.Contains(marker, "supervisor timeout") || !strings.Contains(marker, shared.EnvSupervisorTimeout) {
		t.Fatalf("expected timeout marker with env hint, got %q", marker)
	}
}

func TestClassifyStop_Iterations(t *testing.T) {
	marker, stopped := newTestPipeline().classifyStop(context.Background(), errors.New("max iterations reached: 7"))
	if !stopped || !strings.Contains(marker, "max 7 tool iterations") || !strings.Contains(marker, shared.EnvSupervisorMaxIterations) {
		t.Fatalf("expected iterations marker with env hint, got %q", marker)
	}
}

func TestClassifyStop_RealError(t *testing.T) {
	marker, stopped := newTestPipeline().classifyStop(context.Background(), errors.New("upstream 500"))
	if stopped || marker != "" {
		t.Fatalf("expected not stopped, got stopped=%v marker=%q", stopped, marker)
	}
}
