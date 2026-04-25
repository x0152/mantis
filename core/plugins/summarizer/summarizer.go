package summarizer

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"mantis/core/plugins/tokenizer"
	"mantis/core/protocols"
	"mantis/core/types"
	"mantis/shared"
)

const (
	DefaultCompactTokens = 100000
	DefaultMinKeepRecent = 2
	DefaultMinCompact    = 2
	DefaultTargetRatio   = 0.5
	SummaryToolName      = "compact_context"
	SummaryToolIcon      = "layers"
)

const summaryPromptTpl = `You are maintaining a rolling summary of a long chat between a user and an AI assistant.

Goal: compress the earlier conversation so the assistant can continue without losing critical context, while staying within a strict size budget.

Hard budget: the final summary MUST fit in about %d tokens (~%d characters). Stay at or below this budget. If the budget is tight, drop lower-value bullets rather than overflow.

Keep (in priority order):
1. User identity basics (name, role, location/timezone) — one line
2. Stable facts the assistant will need later: named entities, identifiers, servers, files, IPs, URLs, numbers, versions
3. Goals, constraints, deadlines, budgets
4. Current task state and open questions

Drop:
- Pleasantries, acknowledgements, restated facts
- Meta-commentary about the assistant
- Raw tool outputs, full logs, code dumps

Style:
- Plain text, very dense, short bullets grouped by topic headings
- Third person ("the user asked", "the assistant confirmed")
- Preserve exact identifiers verbatim; do NOT invent details
- Same language as the conversation

Return ONLY the summary text. No preamble, no closing remarks.`

type Input struct {
	SessionID string
	RequestID string
	ModelID   string
	PresetID  string
}

type Result struct {
	Compacted  bool
	Step       *types.Step
	Compressed int
	Before     int
	After      int
}

type MemoryFlusher interface {
	FlushCompactedWindow(ctx context.Context, msgs []types.ChatMessage)
}

type Summarizer struct {
	llm           protocols.LLM
	sessionStore  protocols.Store[string, types.ChatSession]
	messageStore  protocols.Store[string, types.ChatMessage]
	modelStore    protocols.Store[string, types.Model]
	presetStore   protocols.Store[string, types.Preset]
	llmConnStore  protocols.Store[string, types.LlmConnection]
	buffer        *shared.Buffer
	memoryFlusher MemoryFlusher
	minKeepRecent int
	minCompact    int
	targetRatio   float64
}

func New(
	llm protocols.LLM,
	sessionStore protocols.Store[string, types.ChatSession],
	messageStore protocols.Store[string, types.ChatMessage],
	modelStore protocols.Store[string, types.Model],
	presetStore protocols.Store[string, types.Preset],
	llmConnStore protocols.Store[string, types.LlmConnection],
	buffer *shared.Buffer,
) *Summarizer {
	return &Summarizer{
		llm:           llm,
		sessionStore:  sessionStore,
		messageStore:  messageStore,
		modelStore:    modelStore,
		presetStore:   presetStore,
		llmConnStore:  llmConnStore,
		buffer:        buffer,
		minKeepRecent: DefaultMinKeepRecent,
		minCompact:    DefaultMinCompact,
		targetRatio:   DefaultTargetRatio,
	}
}

func (s *Summarizer) SetMemoryFlusher(f MemoryFlusher) {
	if s == nil {
		return
	}
	s.memoryFlusher = f
}

func EffectiveThreshold(model types.Model) int {
	if model.CompactTokens > 0 {
		return model.CompactTokens
	}
	if model.ContextWindow > 0 {
		if budget := model.ContextWindow - model.ReserveTokens; budget > 0 {
			return budget
		}
		return model.ContextWindow
	}
	return DefaultCompactTokens
}

func (s *Summarizer) MaybeCompact(ctx context.Context, in Input) (Result, error) {
	return s.compact(ctx, in, false)
}

func (s *Summarizer) ForceCompact(ctx context.Context, in Input) (Result, error) {
	return s.compact(ctx, in, true)
}

func (s *Summarizer) compact(ctx context.Context, in Input, force bool) (Result, error) {
	if s == nil || s.sessionStore == nil || s.messageStore == nil {
		return Result{}, nil
	}
	sessionID := strings.TrimSpace(in.SessionID)
	if sessionID == "" || strings.HasPrefix(sessionID, "plan:") {
		return Result{}, nil
	}

	model, err := shared.ResolveModel(ctx, s.modelStore, in.ModelID)
	if err != nil {
		return Result{}, nil
	}
	threshold := EffectiveThreshold(model)

	session, err := s.loadSession(ctx, sessionID)
	if err != nil {
		return Result{}, nil
	}

	kept, err := s.loadSessionMessages(ctx, sessionID, session.SummarizedUpTo)
	if err != nil {
		return Result{}, err
	}

	tok := tokenizer.For(model.Name)
	before := estimateConversationTokens(tok, session.SummaryText, kept)
	if !force && before <= threshold {
		return Result{}, nil
	}

	var cutIdx int
	if force {
		cutIdx = forceCut(kept, s.minKeepRecent)
	} else {
		cutIdx = chooseCut(tok, kept, threshold, s.minKeepRecent, s.targetRatio)
	}
	if cutIdx < s.minCompact {
		return Result{}, nil
	}
	toCompact := kept[:cutIdx]
	remaining := kept[cutIdx:]

	step := s.startStep(in.RequestID, len(toCompact), before, force)

	if s.memoryFlusher != nil {
		s.memoryFlusher.FlushCompactedWindow(ctx, toCompact)
	}

	summaryModel, summaryConn, err := s.pickSummaryModel(ctx, in.PresetID, model)
	if err != nil {
		s.finishStepError(in.RequestID, step, err)
		return Result{Step: step}, nil
	}

	budget := threshold / 4
	if budget < 80 {
		budget = 80
	}
	replacedTokens := tok.Count(session.SummaryText) + estimateMessagesTokens(tok, toCompact)

	newSummary, err := s.generateSummary(ctx, summaryConn, summaryModel, session.SummaryText, toCompact, budget)
	if err != nil {
		s.finishStepError(in.RequestID, step, err)
		return Result{Step: step}, nil
	}

	newSummaryTokens := tok.Count(newSummary)
	if !force && newSummaryTokens >= replacedTokens {
		s.finishStepSkipped(in.RequestID, step, len(toCompact), before, replacedTokens, newSummaryTokens)
		return Result{Step: step}, nil
	}

	lastCompacted := toCompact[len(toCompact)-1].CreatedAt
	session.SummaryText = newSummary
	session.SummarizedUpTo = &lastCompacted
	session.SummaryVersion++
	if _, err := s.sessionStore.Update(ctx, []types.ChatSession{session}); err != nil {
		s.finishStepError(in.RequestID, step, err)
		return Result{Step: step}, nil
	}

	after := newSummaryTokens + estimateMessagesTokens(tok, remaining)
	s.finishStepSuccess(in.RequestID, step, len(toCompact), before, after)
	return Result{
		Compacted:  true,
		Step:       step,
		Compressed: len(toCompact),
		Before:     before,
		After:      after,
	}, nil
}

func chooseCut(tok tokenizer.Tokenizer, kept []types.ChatMessage, threshold, minKeepRecent int, targetRatio float64) int {
	target := int(float64(threshold) * targetRatio)
	if target < 1 {
		target = 1
	}
	maxCut := len(kept) - minKeepRecent
	if maxCut < 1 {
		return 0
	}
	acc := 0
	for i := len(kept) - 1; i >= 0; i-- {
		acc += tok.Count(kept[i].Content) + 4
		if acc > target {
			cut := i
			if cut > maxCut {
				cut = maxCut
			}
			return cut
		}
	}
	return 0
}

func forceCut(kept []types.ChatMessage, minKeepRecent int) int {
	cut := len(kept) - minKeepRecent
	if cut < 0 {
		cut = 0
	}
	return cut
}

func (s *Summarizer) loadSession(ctx context.Context, sessionID string) (types.ChatSession, error) {
	sessions, err := s.sessionStore.Get(ctx, []string{sessionID})
	if err != nil {
		return types.ChatSession{}, err
	}
	session, ok := sessions[sessionID]
	if !ok {
		return types.ChatSession{}, fmt.Errorf("session %s not found", sessionID)
	}
	return session, nil
}

func (s *Summarizer) loadSessionMessages(ctx context.Context, sessionID string, after *time.Time) ([]types.ChatMessage, error) {
	all, err := s.messageStore.List(ctx, types.ListQuery{})
	if err != nil {
		return nil, err
	}
	var kept []types.ChatMessage
	for _, m := range all {
		if m.SessionID != sessionID {
			continue
		}
		if m.Status != "" {
			continue
		}
		if after != nil && !m.CreatedAt.After(*after) {
			continue
		}
		kept = append(kept, m)
	}
	sort.Slice(kept, func(i, j int) bool { return kept[i].CreatedAt.Before(kept[j].CreatedAt) })
	return kept, nil
}

func estimateConversationTokens(tok tokenizer.Tokenizer, summary string, msgs []types.ChatMessage) int {
	return tok.Count(summary) + estimateMessagesTokens(tok, msgs)
}

func estimateMessagesTokens(tok tokenizer.Tokenizer, msgs []types.ChatMessage) int {
	total := 0
	for _, m := range msgs {
		total += tok.Count(m.Content) + 4
	}
	return total
}

func (s *Summarizer) pickSummaryModel(ctx context.Context, presetID string, fallback types.Model) (types.Model, types.LlmConnection, error) {
	modelID := fallback.ID
	if presetID != "" && s.presetStore != nil {
		if p, err := shared.ResolvePreset(ctx, s.presetStore, presetID); err == nil {
			if p.SummaryModelID != "" {
				modelID = p.SummaryModelID
			}
		}
	}
	model, err := shared.ResolveModel(ctx, s.modelStore, modelID)
	if err != nil {
		return types.Model{}, types.LlmConnection{}, err
	}
	conn, err := shared.ResolveConnection(ctx, s.llmConnStore, model.ConnectionID)
	if err != nil {
		return types.Model{}, types.LlmConnection{}, err
	}
	return model, conn, nil
}

func (s *Summarizer) generateSummary(ctx context.Context, conn types.LlmConnection, model types.Model, priorSummary string, msgs []types.ChatMessage, budgetTokens int) (string, error) {
	var input strings.Builder
	if priorSummary != "" {
		input.WriteString("Previous summary:\n")
		input.WriteString(priorSummary)
		input.WriteString("\n\n")
	}
	input.WriteString("New conversation chunk to merge into the summary:\n")
	for _, m := range msgs {
		role := m.Role
		if role == "" {
			role = "user"
		}
		input.WriteString(strings.ToUpper(role))
		input.WriteString(": ")
		input.WriteString(m.Content)
		input.WriteString("\n\n")
	}

	systemPrompt := fmt.Sprintf(summaryPromptTpl, budgetTokens, budgetTokens*4)

	messages := []protocols.LLMMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: input.String()},
	}

	stream, err := s.llm.ChatStream(ctx, conn.Provider, conn.BaseURL, conn.APIKey, messages, model.Name, nil, "skip")
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	for ev := range stream {
		if ev.Type == "text" {
			sb.WriteString(ev.Delta)
		}
		if ev.Type == "error" {
			return "", fmt.Errorf("%s", ev.Delta)
		}
	}
	result := strings.TrimSpace(sb.String())
	if result == "" {
		return "", fmt.Errorf("summary model returned empty content")
	}
	return result, nil
}

func (s *Summarizer) startStep(requestID string, count, before int, force bool) *types.Step {
	label := fmt.Sprintf("Compressing earlier context (%d messages, ~%s tokens)…", count, humanTokens(before))
	if force {
		label = fmt.Sprintf("Context overflow — forcing compaction of %d messages (~%s tokens)…", count, humanTokens(before))
	}
	step := &types.Step{
		ID:        "compact-" + uuid.New().String(),
		Tool:      SummaryToolName,
		Icon:      SummaryToolIcon,
		Label:     label,
		Status:    "running",
		StartedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if s.buffer != nil && requestID != "" {
		s.buffer.SetStep(requestID, *step)
	}
	return step
}

func (s *Summarizer) finishStepSuccess(requestID string, step *types.Step, count, before, after int) {
	if step == nil {
		return
	}
	step.Status = "completed"
	step.Result = fmt.Sprintf("Compressed %d messages: ~%s → ~%s tokens", count, humanTokens(before), humanTokens(after))
	step.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	if s.buffer != nil && requestID != "" {
		s.buffer.SetStep(requestID, *step)
	}
}

func (s *Summarizer) finishStepSkipped(requestID string, step *types.Step, count, before, replaced, got int) {
	if step == nil {
		return
	}
	step.Status = "completed"
	step.Label = fmt.Sprintf("Context at ~%s tokens — skipping compaction", humanTokens(before))
	step.Result = fmt.Sprintf("Skipped: summary (~%s tokens) would be larger than the %d messages it replaces (~%s). Nothing to gain from this chunk.", humanTokens(got), count, humanTokens(replaced))
	step.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	if s.buffer != nil && requestID != "" {
		s.buffer.SetStep(requestID, *step)
	}
}

func (s *Summarizer) finishStepError(requestID string, step *types.Step, err error) {
	if step == nil {
		return
	}
	step.Status = "error"
	step.Result = "Compaction failed: " + err.Error()
	step.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	if s.buffer != nil && requestID != "" {
		s.buffer.SetStep(requestID, *step)
	}
}

func humanTokens(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}
