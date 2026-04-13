package plans

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/google/uuid"

	modelplugin "mantis/core/plugins/model"
	sessionplugin "mantis/core/plugins/session"
	"mantis/core/protocols"
	"mantis/core/types"
	messageworkflow "mantis/core/workflows/message"
	"mantis/shared"
)

const maxTransitions = 50

type Runner struct {
	planStore     protocols.Store[string, types.Plan]
	runStore      protocols.Store[string, types.PlanRun]
	messageStore  protocols.Store[string, types.ChatMessage]
	sessionPolicy *sessionplugin.Policy
	workflow      *messageworkflow.Workflow
	buffer        *shared.Buffer

	mu      sync.Mutex
	cancels map[string]context.CancelFunc
	done    map[string]chan struct{}
}

func NewRunner(
	planStore protocols.Store[string, types.Plan],
	runStore protocols.Store[string, types.PlanRun],
	messageStore protocols.Store[string, types.ChatMessage],
	sessionPolicy *sessionplugin.Policy,
	workflow *messageworkflow.Workflow,
	buffer *shared.Buffer,
) *Runner {
	return &Runner{
		planStore:     planStore,
		runStore:      runStore,
		messageStore:  messageStore,
		sessionPolicy: sessionPolicy,
		workflow:      workflow,
		buffer:        buffer,
		cancels:       make(map[string]context.CancelFunc),
		done:          make(map[string]chan struct{}),
	}
}

func (r *Runner) TriggerRun(ctx context.Context, planID, trigger string, input map[string]any) (types.PlanRun, error) {
	plans, err := r.planStore.Get(ctx, []string{planID})
	if err != nil {
		return types.PlanRun{}, err
	}
	plan, ok := plans[planID]
	if !ok {
		return types.PlanRun{}, fmt.Errorf("plan not found: %s", planID)
	}

	if err := validateGraph(plan.Graph); err != nil {
		return types.PlanRun{}, fmt.Errorf("invalid graph: %w", err)
	}

	if input == nil {
		input = map[string]any{}
	}

	now := time.Now().UTC()
	run := types.PlanRun{
		ID:        uuid.New().String(),
		PlanID:    planID,
		Status:    "running",
		Trigger:   trigger,
		Input:     input,
		Steps:     initSteps(plan.Graph),
		StartedAt: now,
	}

	created, err := r.runStore.Create(ctx, []types.PlanRun{run})
	if err != nil {
		return types.PlanRun{}, err
	}
	run = created[0]

	go r.execute(context.Background(), plan, run)

	return run, nil
}

func (r *Runner) CancelRun(ctx context.Context, runID string) (types.PlanRun, error) {
	r.mu.Lock()
	cancel, inMemory := r.cancels[runID]
	doneCh := r.done[runID]
	r.mu.Unlock()

	if inMemory {
		cancel()
		select {
		case <-doneCh:
		case <-time.After(5 * time.Second):
		}
	}

	runs, err := r.runStore.Get(ctx, []string{runID})
	if err != nil {
		return types.PlanRun{}, err
	}
	run, ok := runs[runID]
	if !ok {
		return types.PlanRun{}, fmt.Errorf("run not found: %s", runID)
	}

	if run.Status == "running" {
		now := time.Now().UTC()
		run.Status = "cancelled"
		run.FinishedAt = &now
		markStaleSteps(&run, now, "cancelled")
		if _, err := r.runStore.Update(ctx, []types.PlanRun{run}); err != nil {
			return types.PlanRun{}, err
		}
	}

	return run, nil
}

func (r *Runner) ActiveRuns(ctx context.Context) ([]types.PlanRun, error) {
	runs, err := r.runStore.List(ctx, types.ListQuery{
		Filter: map[string]string{"status": "running"},
	})
	if err != nil {
		return nil, err
	}
	if runs == nil {
		runs = []types.PlanRun{}
	}
	return runs, nil
}

func (r *Runner) RecoverStaleRuns(ctx context.Context) {
	runs, err := r.runStore.List(ctx, types.ListQuery{
		Filter: map[string]string{"status": "running"},
	})
	if err != nil {
		log.Printf("plans: recover stale runs: %v", err)
		return
	}
	now := time.Now().UTC()
	for _, run := range runs {
		run.Status = "failed"
		run.FinishedAt = &now
		markStaleSteps(&run, now, "interrupted by server restart")
		if _, err := r.runStore.Update(ctx, []types.PlanRun{run}); err != nil {
			log.Printf("plans: recover run %s: %v", run.ID, err)
		} else {
			log.Printf("plans: recovered stale run %s (marked as failed)", run.ID)
		}
	}
}

func (r *Runner) execute(ctx context.Context, plan types.Plan, run types.PlanRun) {
	ctx, cancel := context.WithCancel(ctx)
	doneCh := make(chan struct{})
	r.mu.Lock()
	r.cancels[run.ID] = cancel
	r.done[run.ID] = doneCh
	r.mu.Unlock()
	defer func() {
		r.mu.Lock()
		delete(r.cancels, run.ID)
		delete(r.done, run.ID)
		r.mu.Unlock()
		cancel()
		close(doneCh)
	}()

	sessionID := fmt.Sprintf("plan:%s:%s", plan.ID, run.ID)

	if r.buffer != nil {
		r.buffer.MarkSessionActive(sessionID)
		defer r.buffer.MarkSessionInactive(sessionID)
	}

	if _, err := r.sessionPolicy.Execute(ctx, sessionplugin.Input{
		Mode:      sessionplugin.ModeEnsure,
		SessionID: sessionID,
		Source:    "plan",
		Title:     fmt.Sprintf("Plan: %s", plan.Name),
	}); err != nil {
		log.Printf("plans: ensure session: %v", err)
		r.finishRun(&run, "failed")
		return
	}

	startNodes := findStartNodes(plan.Graph)
	if len(startNodes) == 0 {
		r.finishRun(&run, "failed")
		return
	}

	nodeMap := make(map[string]types.PlanNode, len(plan.Graph.Nodes))
	for _, n := range plan.Graph.Nodes {
		nodeMap[n.ID] = n
	}

	transitions := 0
	current := startNodes[0]
	for {
		transitions++
		if transitions > maxTransitions {
			r.failStep(&run, current, fmt.Sprintf("max transitions exceeded (%d)", maxTransitions))
			r.finishRun(&run, "failed")
			return
		}

		if ctx.Err() != nil {
			r.failStep(&run, current, "cancelled")
			r.finishRun(&run, "cancelled")
			return
		}

		node, ok := nodeMap[current]
		if !ok {
			r.failStep(&run, current, "node not found in graph")
			r.finishRun(&run, "failed")
			return
		}

		r.markStepRunning(&run, current)

		prompt := renderPrompt(node.Prompt, run.Input)
		if node.Type == types.PlanNodeDecision {
			prompt = decisionPrompt(node, run.Input)
		}

		res, err := r.executeWithRetry(ctx, sessionID, node, prompt)
		if err != nil {
			if ctx.Err() != nil {
				r.failStep(&run, current, "cancelled")
				r.finishRun(&run, "cancelled")
				return
			}
			r.failStep(&run, current, err.Error())
			r.finishRun(&run, "failed")
			return
		}

		switch node.Type {
		case types.PlanNodeAction:
			r.completeStep(&run, current, res.messageID)

		case types.PlanNodeDecision:
			branch := parseDecision(res.content)
			r.setStepResult(&run, current, "completed", branch)

			next := findEdgeTarget(plan.Graph, current, branch)
			if next == "" {
				skipPending(&run)
				r.saveRun(&run)
				r.finishRun(&run, "completed")
				return
			}
			current = next
			continue

		default:
			r.failStep(&run, current, fmt.Sprintf("unsupported node type: %s", node.Type))
			r.finishRun(&run, "failed")
			return
		}

		next := findNextNode(plan.Graph, current)
		if next == "" {
			skipPending(&run)
			r.saveRun(&run)
			r.finishRun(&run, "completed")
			return
		}
		current = next
	}
}

func (r *Runner) executeWithRetry(ctx context.Context, sessionID string, node types.PlanNode, prompt string) (nodeResult, error) {
	maxRetries := node.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("plans: retrying node %s (attempt %d/%d)", node.ID, attempt+1, maxRetries+1)
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		if ctx.Err() != nil {
			return nodeResult{}, ctx.Err()
		}
		res, err := r.executeNode(ctx, sessionID, prompt, node.ClearContext)
		if err == nil {
			return res, nil
		}
		lastErr = err
	}
	return nodeResult{}, lastErr
}

type nodeResult struct {
	content   string
	messageID string
}

func (r *Runner) executeNode(ctx context.Context, sessionID, prompt string, clearContext bool) (nodeResult, error) {
	done := make(chan struct{})
	out, err := r.workflow.Execute(ctx, messageworkflow.Input{
		SessionID:      sessionID,
		Content:        prompt,
		Source:         "plan",
		ModelConfig:    modelplugin.Input{DefaultPreset: "chat"},
		DisableHistory: clearContext,
		Timeout:        10 * time.Minute,
		Finally:        func() { close(done) },
	})
	if err != nil {
		return nodeResult{}, err
	}

	select {
	case <-done:
	case <-ctx.Done():
		return nodeResult{}, ctx.Err()
	}

	messages, err := r.messageStore.Get(ctx, []string{out.AssistantMessage.ID})
	if err != nil {
		return nodeResult{}, err
	}
	msg, ok := messages[out.AssistantMessage.ID]
	if !ok {
		return nodeResult{}, fmt.Errorf("assistant message not found")
	}
	if msg.Status == "error" {
		return nodeResult{messageID: msg.ID}, fmt.Errorf("step error: %s", msg.Content)
	}
	if strings.Contains(msg.Content, "[ERROR]:") {
		return nodeResult{messageID: msg.ID}, fmt.Errorf("step reported error: %s", msg.Content)
	}
	return nodeResult{content: msg.Content, messageID: msg.ID}, nil
}

func renderPrompt(raw string, input map[string]any) string {
	if !strings.Contains(raw, "{{") {
		return raw
	}
	if input == nil {
		input = map[string]any{}
	}
	tmpl, err := template.New("prompt").Option("missingkey=zero").Parse(raw)
	if err != nil {
		return raw
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, input); err != nil {
		return raw
	}
	return buf.String()
}

func decisionPrompt(node types.PlanNode, input map[string]any) string {
	rendered := renderPrompt(node.Prompt, input)
	if node.ClearContext {
		return fmt.Sprintf(
			"Answer this question. Reply with EXACTLY 'yes' or 'no' as the FIRST word of your response, then explain briefly.\n\nQuestion: %s",
			rendered,
		)
	}
	return fmt.Sprintf(
		"Based on everything above, answer this question. Reply with EXACTLY 'yes' or 'no' as the FIRST word of your response, then explain briefly.\n\nQuestion: %s",
		rendered,
	)
}

func parseDecision(response string) string {
	fields := strings.Fields(strings.TrimSpace(strings.ToLower(response)))
	if len(fields) == 0 {
		return "yes"
	}
	first := strings.TrimRight(fields[0], ".,!:;")
	if strings.HasPrefix(first, "no") {
		return "no"
	}
	return "yes"
}

func (r *Runner) markStepRunning(run *types.PlanRun, nodeID string) {
	now := time.Now().UTC()
	for i := range run.Steps {
		if run.Steps[i].NodeID == nodeID {
			run.Steps[i].Status = "running"
			if run.Steps[i].StartedAt == nil {
				run.Steps[i].StartedAt = &now
			}
			break
		}
	}
	r.saveRun(run)
}

func (r *Runner) completeStep(run *types.PlanRun, nodeID, messageID string) {
	now := time.Now().UTC()
	for i := range run.Steps {
		if run.Steps[i].NodeID == nodeID {
			run.Steps[i].Status = "completed"
			run.Steps[i].MessageID = messageID
			run.Steps[i].FinishedAt = &now
			break
		}
	}
	r.saveRun(run)
}

func (r *Runner) failStep(run *types.PlanRun, nodeID, result string) {
	now := time.Now().UTC()
	for i := range run.Steps {
		if run.Steps[i].NodeID == nodeID {
			run.Steps[i].Status = "failed"
			run.Steps[i].Result = result
			run.Steps[i].FinishedAt = &now
			break
		}
	}
	r.saveRun(run)
}

func (r *Runner) setStepResult(run *types.PlanRun, nodeID, status, result string) {
	now := time.Now().UTC()
	for i := range run.Steps {
		if run.Steps[i].NodeID == nodeID {
			run.Steps[i].Status = status
			run.Steps[i].Result = result
			run.Steps[i].FinishedAt = &now
			break
		}
	}
	r.saveRun(run)
}

func (r *Runner) finishRun(run *types.PlanRun, status string) {
	now := time.Now().UTC()
	run.Status = status
	run.FinishedAt = &now
	r.saveRun(run)
	log.Printf("plans: run %s finished with status=%s", run.ID, status)
}

func (r *Runner) saveRun(run *types.PlanRun) {
	if _, err := r.runStore.Update(context.Background(), []types.PlanRun{*run}); err != nil {
		log.Printf("plans: save run %s: %v", run.ID, err)
	}
}

func markStaleSteps(run *types.PlanRun, now time.Time, reason string) {
	for i := range run.Steps {
		switch run.Steps[i].Status {
		case "running":
			run.Steps[i].Status = "failed"
			run.Steps[i].Result = reason
			run.Steps[i].FinishedAt = &now
		case "pending":
			run.Steps[i].Status = "skipped"
		}
	}
}

func initSteps(graph types.PlanGraph) []types.PlanStepRun {
	steps := make([]types.PlanStepRun, len(graph.Nodes))
	for i, n := range graph.Nodes {
		steps[i] = types.PlanStepRun{NodeID: n.ID, Status: "pending"}
	}
	return steps
}

func skipPending(run *types.PlanRun) {
	for i := range run.Steps {
		if run.Steps[i].Status == "pending" {
			run.Steps[i].Status = "skipped"
		}
	}
}

func validateGraph(graph types.PlanGraph) error {
	nodeTypes := make(map[string]types.PlanNodeType, len(graph.Nodes))
	for _, n := range graph.Nodes {
		nodeTypes[n.ID] = n.Type
	}
	outCount := make(map[string]int, len(graph.Edges))
	for _, e := range graph.Edges {
		outCount[e.Source]++
	}
	for nodeID, count := range outCount {
		nt := nodeTypes[nodeID]
		if nt == types.PlanNodeAction && count > 1 {
			return fmt.Errorf("action node %q has %d outgoing edges (max 1)", nodeID, count)
		}
		if nt == types.PlanNodeDecision && count > 2 {
			return fmt.Errorf("decision node %q has %d outgoing edges (max 2)", nodeID, count)
		}
	}
	return nil
}

func findStartNodes(graph types.PlanGraph) []string {
	targets := make(map[string]bool, len(graph.Edges))
	for _, e := range graph.Edges {
		targets[e.Target] = true
	}
	var starts []string
	for _, n := range graph.Nodes {
		if !targets[n.ID] {
			starts = append(starts, n.ID)
		}
	}
	return starts
}

func findNextNode(graph types.PlanGraph, fromNodeID string) string {
	for _, e := range graph.Edges {
		if e.Source == fromNodeID {
			return e.Target
		}
	}
	return ""
}

func findEdgeTarget(graph types.PlanGraph, fromNodeID, label string) string {
	label = strings.TrimSpace(strings.ToLower(label))
	for _, e := range graph.Edges {
		if e.Source == fromNodeID && strings.TrimSpace(strings.ToLower(e.Label)) == label {
			return e.Target
		}
	}
	for _, e := range graph.Edges {
		if e.Source == fromNodeID {
			return e.Target
		}
	}
	return ""
}
