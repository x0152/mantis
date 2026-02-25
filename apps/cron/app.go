package cron

import (
	"context"
	"log"
	"sync"
	"time"

	robcron "github.com/robfig/cron/v3"

	usecases "mantis/apps/cron/use_cases"
	"mantis/core/agents"
	artifactplugin "mantis/core/plugins/artifact"
	modelplugin "mantis/core/plugins/model"
	"mantis/core/plugins/pipeline"
	sessionplugin "mantis/core/plugins/session"
	"mantis/core/protocols"
	"mantis/core/types"
	messageworkflow "mantis/core/workflows/message"
)

type App struct {
	ucSyncJobs   *usecases.SyncJobs
	ucExecuteJob *usecases.ExecuteJob

	mu       sync.Mutex
	sched    *robcron.Cron
	entries  map[string]robcron.EntryID
	parser   robcron.Parser
	syncFreq time.Duration
}

func NewApp(
	configStore protocols.Store[string, types.Config],
	channelStore protocols.Store[string, types.Channel],
	sessionStore protocols.Store[string, types.ChatSession],
	messageStore protocols.Store[string, types.ChatMessage],
	modelStore protocols.Store[string, types.Model],
	cronJobStore protocols.Store[string, types.CronJob],
	agent *agents.MantisAgent,
	artifactMgr *artifactplugin.Manager,
	memoryExtractor pipeline.MemoryExtractor,
) *App {
	if artifactMgr == nil {
		artifactMgr = artifactplugin.NewManager(nil)
	}
	parser := robcron.NewParser(robcron.Minute | robcron.Hour | robcron.Dom | robcron.Month | robcron.Dow | robcron.Descriptor)

	modelResolver := modelplugin.NewResolver(nil, configStore)
	workflow := messageworkflow.New(messageStore, modelStore, agent, nil, modelResolver, artifactMgr, memoryExtractor)

	sessionPolicy := sessionplugin.NewPolicy(sessionStore)
	loadConfigUC := usecases.NewLoadConfig(configStore)
	ensureSessionUC := usecases.NewEnsureSession(sessionPolicy)
	executeJobUC := usecases.NewExecuteJob(loadConfigUC, ensureSessionUC, channelStore, workflow)

	return &App{
		ucSyncJobs:   usecases.NewSyncJobs(cronJobStore),
		ucExecuteJob: executeJobUC,
		entries:      make(map[string]robcron.EntryID),
		parser:       parser,
		syncFreq:     30 * time.Second,
	}
}

func (a *App) Start(ctx context.Context) {
	a.mu.Lock()
	a.sched = robcron.New(robcron.WithParser(a.parser))
	a.mu.Unlock()

	a.syncJobs(ctx)
	a.sched.Start()
	log.Printf("cron: scheduler started")

	ticker := time.NewTicker(a.syncFreq)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.mu.Lock()
			if a.sched != nil {
				a.sched.Stop()
			}
			a.mu.Unlock()
			log.Printf("cron: scheduler stopped: %v", ctx.Err())
			return
		case <-ticker.C:
			a.syncJobs(ctx)
		}
	}
}
