package usecases

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	modelplugin "mantis/core/plugins/model"
	"mantis/core/protocols"
	"mantis/core/types"
	messageworkflow "mantis/core/workflows/message"
	adapter "mantis/infrastructure/adapters/channel"
)

type ExecuteJobOutput struct {
	Skipped bool
}

type ExecuteJob struct {
	loadConfig    *LoadConfig
	ensureSession *EnsureSession
	channelStore  protocols.Store[string, types.Channel]
	workflow      *messageworkflow.Workflow

	mu      sync.Mutex
	running map[string]bool
}

func NewExecuteJob(
	loadConfig *LoadConfig,
	ensureSession *EnsureSession,
	channelStore protocols.Store[string, types.Channel],
	workflow *messageworkflow.Workflow,
) *ExecuteJob {
	return &ExecuteJob{
		loadConfig:    loadConfig,
		ensureSession: ensureSession,
		channelStore:  channelStore,
		workflow:      workflow,
		running:       make(map[string]bool),
	}
}

func (uc *ExecuteJob) Execute(ctx context.Context, job types.CronJob) (ExecuteJobOutput, error) {
	if !uc.markRunning(job.ID) {
		return ExecuteJobOutput{Skipped: true}, nil
	}

	cfg, err := uc.loadConfig.Execute(ctx)
	if err != nil {
		uc.unmarkRunning(job.ID)
		return ExecuteJobOutput{}, err
	}

	sessionID := "cron:" + job.ID
	if err := uc.ensureSession.Execute(ctx, sessionID); err != nil {
		uc.unmarkRunning(job.ID)
		return ExecuteJobOutput{}, err
	}

	sender, err := uc.resolveSender(ctx, cfg.Cron.Channel, cfg.Cron.Sender)
	if err != nil {
		uc.unmarkRunning(job.ID)
		return ExecuteJobOutput{}, err
	}

	_, err = uc.workflow.Execute(ctx, messageworkflow.Input{
		SessionID:      sessionID,
		Content:        job.Prompt,
		Source:         "cron",
		ResponseTo:     sender,
		ModelConfig:    modelplugin.Input{ConfigPath: []string{"cron", "model_id"}},
		DisableHistory: true,
		ErrorPrefix:    "[Error]",
		Timeout:        5 * time.Minute,
		Finally:        func() { uc.unmarkRunning(job.ID) },
	})
	if err != nil {
		uc.unmarkRunning(job.ID)
		return ExecuteJobOutput{}, err
	}

	return ExecuteJobOutput{}, nil
}

func (uc *ExecuteJob) resolveSender(ctx context.Context, channel, recipient string) (protocols.ResponseTo, error) {
	channel = strings.TrimSpace(strings.ToLower(channel))
	if channel == "" {
		return nil, nil
	}
	switch channel {
	case "telegram":
		token, err := uc.telegramToken(ctx)
		if err != nil {
			return nil, err
		}
		if token == "" {
			return nil, fmt.Errorf("telegram channel configured but no token found")
		}
		return adapter.NewTelegramResponseTo(token, recipient), nil
	default:
		return nil, fmt.Errorf("unsupported delivery channel %q", channel)
	}
}

func (uc *ExecuteJob) telegramToken(ctx context.Context) (string, error) {
	if uc.channelStore == nil {
		return "", nil
	}
	channels, err := uc.channelStore.List(ctx, types.ListQuery{})
	if err != nil {
		return "", err
	}
	for _, ch := range channels {
		if ch.Type == "telegram" && ch.Token != "" {
			return ch.Token, nil
		}
	}
	return "", nil
}

func (uc *ExecuteJob) markRunning(jobID string) bool {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if uc.running[jobID] {
		return false
	}
	uc.running[jobID] = true
	return true
}

func (uc *ExecuteJob) unmarkRunning(jobID string) {
	uc.mu.Lock()
	delete(uc.running, jobID)
	uc.mu.Unlock()
}
