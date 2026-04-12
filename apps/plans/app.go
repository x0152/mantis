package plans

import (
	"context"
	"log"
	"sync"
	"time"

	robcron "github.com/robfig/cron/v3"

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
	runner *Runner

	planStore protocols.Store[string, types.Plan]

	mu       sync.Mutex
	sched    *robcron.Cron
	entries  map[string]robcron.EntryID
	parser   robcron.Parser
	syncFreq time.Duration
}

func NewApp(
	settingsStore protocols.Store[string, types.Settings],
	sessionStore protocols.Store[string, types.ChatSession],
	messageStore protocols.Store[string, types.ChatMessage],
	modelStore protocols.Store[string, types.Model],
	presetStore protocols.Store[string, types.Preset],
	planStore protocols.Store[string, types.Plan],
	runStore protocols.Store[string, types.PlanRun],
	agent *agents.MantisAgent,
	artifactMgr *artifactplugin.Manager,
	memoryExtractor pipeline.MemoryExtractor,
) *App {
	if artifactMgr == nil {
		artifactMgr = artifactplugin.NewManager(nil)
	}

	modelResolver := modelplugin.NewResolver(nil, settingsStore, presetStore)
	workflow := messageworkflow.New(messageStore, modelStore, agent, nil, modelResolver, artifactMgr, memoryExtractor)
	sessionPolicy := sessionplugin.NewPolicy(sessionStore)
	parser := robcron.NewParser(robcron.Minute | robcron.Hour | robcron.Dom | robcron.Month | robcron.Dow | robcron.Descriptor)

	return &App{
		runner:    NewRunner(planStore, runStore, messageStore, sessionPolicy, workflow),
		planStore: planStore,
		entries:   make(map[string]robcron.EntryID),
		parser:    parser,
		syncFreq:  30 * time.Second,
	}
}

func (a *App) Runner() *Runner {
	return a.runner
}

func (a *App) Start(ctx context.Context) {
	a.runner.RecoverStaleRuns(ctx)

	a.mu.Lock()
	a.sched = robcron.New(robcron.WithParser(a.parser))
	a.mu.Unlock()

	a.syncPlans(ctx)
	a.sched.Start()
	log.Printf("plans: scheduler started")

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
			log.Printf("plans: scheduler stopped: %v", ctx.Err())
			return
		case <-ticker.C:
			a.syncPlans(ctx)
		}
	}
}

func (a *App) syncPlans(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.sched == nil {
		return
	}

	allPlans, err := a.planStore.List(ctx, types.ListQuery{})
	if err != nil {
		log.Printf("plans: sync: %v", err)
		return
	}

	for _, eid := range a.entries {
		a.sched.Remove(eid)
	}
	next := make(map[string]robcron.EntryID)

	for _, p := range allPlans {
		if !p.Enabled || p.Schedule == "" {
			continue
		}
		plan := p
		eid, err := a.sched.AddFunc(plan.Schedule, func() {
			a.executePlan(context.Background(), plan)
		})
		if err != nil {
			log.Printf("plans: invalid schedule for %s: %v", plan.ID, err)
			continue
		}
		next[plan.ID] = eid
	}
	a.entries = next
}

func (a *App) executePlan(ctx context.Context, plan types.Plan) {
	log.Printf("plans: triggering scheduled run for plan=%s (%s)", plan.ID, plan.Name)
	_, err := a.runner.TriggerRun(ctx, plan.ID, "schedule")
	if err != nil {
		log.Printf("plans: trigger run for plan=%s: %v", plan.ID, err)
	}
}
