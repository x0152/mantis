package usecases

import (
	"context"
	"fmt"
	"strconv"
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

type DeliveryConfig struct {
	Channel string
}

type ExecuteJob struct {
	delivery      DeliveryConfig
	ensureSession *EnsureSession
	channelStore  protocols.Store[string, types.Channel]
	workflow      *messageworkflow.Workflow

	mu      sync.Mutex
	running map[string]bool
}

func NewExecuteJob(
	delivery DeliveryConfig,
	ensureSession *EnsureSession,
	channelStore protocols.Store[string, types.Channel],
	workflow *messageworkflow.Workflow,
) *ExecuteJob {
	return &ExecuteJob{
		delivery:      delivery,
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

	sessionID := "cron:" + job.ID
	if err := uc.ensureSession.Execute(ctx, sessionID); err != nil {
		uc.unmarkRunning(job.ID)
		return ExecuteJobOutput{}, err
	}

	sender, err := uc.resolveSender(ctx, uc.delivery.Channel)
	if err != nil {
		uc.unmarkRunning(job.ID)
		return ExecuteJobOutput{}, err
	}

	_, err = uc.workflow.Execute(ctx, messageworkflow.Input{
		SessionID:      sessionID,
		Content:        job.Prompt,
		Source:         "cron",
		ResponseTo:     sender,
		ModelConfig:    modelplugin.Input{DefaultPreset: "chat"},
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

func (uc *ExecuteJob) resolveSender(ctx context.Context, channel string) (protocols.ResponseTo, error) {
	channel = strings.TrimSpace(strings.ToLower(channel))

	switch channel {
	case "":
		return nil, nil
	case "telegram":
		token, recipient, err := uc.telegramDelivery(ctx)
		if err != nil {
			return nil, err
		}
		if token == "" {
			return nil, fmt.Errorf("telegram channel configured but no token found")
		}
		recipient = strings.TrimSpace(recipient)
		if recipient == "" {
			return nil, fmt.Errorf("telegram delivery recipient is empty (set allowedUserIds in telegram channel)")
		}
		return adapter.NewTelegramResponseTo(token, recipient), nil
	default:
		return nil, fmt.Errorf("unsupported delivery channel %q", channel)
	}
}

func (uc *ExecuteJob) telegramDelivery(ctx context.Context) (string, string, error) {
	if uc.channelStore == nil {
		return "", "", nil
	}
	channels, err := uc.channelStore.List(ctx, types.ListQuery{})
	if err != nil {
		return "", "", err
	}
	token := ""
	for _, ch := range channels {
		if ch.Type != "telegram" || strings.TrimSpace(ch.Token) == "" {
			continue
		}
		if token == "" {
			token = strings.TrimSpace(ch.Token)
		}
		if len(ch.AllowedUserIDs) > 0 {
			return token, strconv.FormatInt(ch.AllowedUserIDs[0], 10), nil
		}
	}
	return token, "", nil
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
