package shared

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	EnvSupervisorMaxIterations = "MANTIS_SUPERVISOR_MAX_ITERATIONS"
	EnvSupervisorTimeout       = "MANTIS_SUPERVISOR_TIMEOUT"
	EnvServerMaxIterations     = "MANTIS_SERVER_MAX_ITERATIONS"
	EnvServerTimeout           = "MANTIS_SERVER_TIMEOUT"
	EnvPlanStepTimeout         = "MANTIS_PLAN_STEP_TIMEOUT"
)

type Limits struct {
	SupervisorMaxIterations int
	SupervisorTimeout       time.Duration
	ServerMaxIterations     int
	ServerTimeout           time.Duration
	PlanStepTimeout         time.Duration
}

func DefaultLimits() Limits {
	return Limits{
		SupervisorMaxIterations: 30,
		SupervisorTimeout:       5 * time.Minute,
		ServerMaxIterations:     30,
		ServerTimeout:           5 * time.Minute,
		PlanStepTimeout:         10 * time.Minute,
	}
}

func LoadLimits() Limits {
	l := DefaultLimits()
	l.SupervisorMaxIterations = envInt(EnvSupervisorMaxIterations, l.SupervisorMaxIterations)
	l.SupervisorTimeout = envDuration(EnvSupervisorTimeout, l.SupervisorTimeout)
	l.ServerMaxIterations = envInt(EnvServerMaxIterations, l.ServerMaxIterations)
	l.ServerTimeout = envDuration(EnvServerTimeout, l.ServerTimeout)
	l.PlanStepTimeout = envDuration(EnvPlanStepTimeout, l.PlanStepTimeout)
	return l
}

func envInt(key string, def int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return def
	}
	return v
}

func envDuration(key string, def time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return def
	}
	return d
}

func FormatDuration(d time.Duration) string {
	if d <= 0 {
		return "unlimited"
	}
	return d.String()
}

func StopReasonUser() string {
	return "[stopped by user]"
}

func StopReasonSupervisorTimeout(limit time.Duration) string {
	return fmt.Sprintf("[stopped: supervisor timeout %s exceeded — raise %s in .env to increase]", FormatDuration(limit), EnvSupervisorTimeout)
}

func StopReasonSupervisorIterations(limit int) string {
	return fmt.Sprintf("[stopped: supervisor reached max %d tool iterations — raise %s in .env to increase]", limit, EnvSupervisorMaxIterations)
}

func StopReasonServerTimeout(limit time.Duration) string {
	return fmt.Sprintf("[server call stopped: timeout %s exceeded — raise %s in .env to increase]", FormatDuration(limit), EnvServerTimeout)
}

func StopReasonServerIterations(limit int) string {
	return fmt.Sprintf("[server call stopped: reached max %d tool iterations — raise %s in .env to increase]", limit, EnvServerMaxIterations)
}

func StopReasonPlanStepTimeout(limit time.Duration) string {
	return fmt.Sprintf("[plan step stopped: timeout %s exceeded — raise %s in .env to increase]", FormatDuration(limit), EnvPlanStepTimeout)
}
