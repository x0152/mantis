package agents

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"mantis/shared"
)

func testLimits() shared.Limits {
	return shared.Limits{ServerMaxIterations: 4, ServerTimeout: 2 * time.Second}
}

func TestAnnotateServerLimit_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond)
	defer cancel()
	time.Sleep(2 * time.Millisecond)
	got := annotateServerLimit("partial", errors.New("context deadline exceeded"), ctx, testLimits())
	if !strings.Contains(got, "server call stopped") || !strings.Contains(got, shared.EnvServerTimeout) || !strings.Contains(got, "partial") {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestAnnotateServerLimit_Iterations(t *testing.T) {
	got := annotateServerLimit("", errors.New("max iterations reached: 4"), context.Background(), testLimits())
	if !strings.Contains(got, "max 4 tool iterations") || !strings.Contains(got, shared.EnvServerMaxIterations) {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestAnnotateServerLimit_UserCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got := annotateServerLimit("done partially", errors.New("canceled"), ctx, testLimits())
	if !strings.Contains(got, "stopped by user") || !strings.Contains(got, "done partially") {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestAnnotateServerLimit_GenericError(t *testing.T) {
	got := annotateServerLimit("partial", errors.New("upstream boom"), context.Background(), testLimits())
	if !strings.Contains(got, "upstream boom") || !strings.Contains(got, "partial") {
		t.Fatalf("unexpected: %q", got)
	}
}
