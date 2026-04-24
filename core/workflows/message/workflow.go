package message

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"mantis/core/agents"
	artifactplugin "mantis/core/plugins/artifact"
	modelplugin "mantis/core/plugins/model"
	"mantis/core/plugins/pipeline"
	"mantis/core/plugins/summarizer"
	"mantis/core/protocols"
	"mantis/core/types"
	"mantis/shared"
)

type Input struct {
	SessionID      string
	Content        string
	Incoming       []protocols.FileAttachment
	ModelConfig    modelplugin.Input
	ResponseTo     protocols.ResponseTo
	Source         string
	DisableHistory bool
	ErrorPrefix    string
	Timeout        time.Duration
	Finally        func()
}

type Output struct {
	UserMessage      types.ChatMessage
	AssistantMessage types.ChatMessage
}

type Workflow struct {
	pipeline      *pipeline.RequestHandlePipeline
	messageStore  protocols.Store[string, types.ChatMessage]
	artifactMgr   *artifactplugin.Manager
	cancellations *pipeline.Cancellations
}

func New(
	messageStore protocols.Store[string, types.ChatMessage],
	modelStore protocols.Store[string, types.Model],
	agent *agents.MantisAgent,
	buffer *shared.Buffer,
	modelResolver *modelplugin.Resolver,
	artifactMgr *artifactplugin.Manager,
	memoryExtractor pipeline.MemoryExtractor,
	summ *summarizer.Summarizer,
	cancellations *pipeline.Cancellations,
) *Workflow {
	if artifactMgr == nil {
		artifactMgr = artifactplugin.NewManager(nil)
	}
	return &Workflow{
		pipeline:      pipeline.New(agent, buffer, messageStore, modelStore, modelResolver, memoryExtractor, summ, agent.Limits()),
		messageStore:  messageStore,
		artifactMgr:   artifactMgr,
		cancellations: cancellations,
	}
}

func (w *Workflow) SetAttachmentDir(dir string) {
	w.pipeline.SetAttachmentDir(dir)
}

type RegenerateInput struct {
	SessionID    string
	UserContent  string
	Source       string
	ModelConfig  modelplugin.Input
	Timeout      time.Duration
	ErrorPrefix  string
	ResponseTo   protocols.ResponseTo
}

func (w *Workflow) Regenerate(ctx context.Context, in RegenerateInput) (types.ChatMessage, error) {
	sessionID := strings.TrimSpace(in.SessionID)
	if sessionID == "" {
		return types.ChatMessage{}, fmt.Errorf("session_id is required")
	}

	now := time.Now()
	assistantMsg := types.ChatMessage{
		ID: uuid.New().String(), SessionID: sessionID,
		Role: "assistant", Content: "", Status: "pending",
		Source: in.Source, CreatedAt: now,
	}
	if _, err := w.messageStore.Create(ctx, []types.ChatMessage{assistantMsg}); err != nil {
		return types.ChatMessage{}, err
	}

	artifacts := w.artifactMgr.ForSession(sessionID)

	execCtx, release := w.cancellations.Begin(context.Background(), sessionID)
	go func() {
		defer release()
		_ = w.pipeline.Execute(execCtx, pipeline.Input{
			Message:     assistantMsg,
			SessionID:   sessionID,
			Content:     in.UserContent,
			Artifacts:   artifacts,
			ModelConfig: in.ModelConfig,
			ResponseTo:  in.ResponseTo,
			Source:      in.Source,
			ErrorPrefix: in.ErrorPrefix,
			Timeout:     in.Timeout,
		})
	}()

	return assistantMsg, nil
}

func (w *Workflow) Execute(ctx context.Context, in Input) (Output, error) {
	sessionID := strings.TrimSpace(in.SessionID)
	if sessionID == "" {
		return Output{}, fmt.Errorf("session_id is required")
	}

	artifacts := w.artifactMgr.ForSession(sessionID)
	var uploaded []shared.ArtifactMeta
	for _, f := range in.Incoming {
		if len(f.Data) == 0 {
			continue
		}
		name := f.FileName
		if name == "" {
			name = "attachment"
		}
		meta, err := artifacts.Put(name, f.Data, f.MimeType)
		if err == nil {
			uploaded = append(uploaded, meta)
		}
	}

	content := in.Content
	if len(uploaded) > 0 {
		var sb strings.Builder
		sb.WriteString(content)
		sb.WriteString("\n\nAttached files:")
		for _, m := range uploaded {
			sb.WriteString(fmt.Sprintf("\n- %s (artifact_id=%s, format=%s, %d bytes)", m.Name, m.ID, m.Format, m.SizeBytes))
		}
		content = sb.String()
	}

	now := time.Now()
	userMsg := types.ChatMessage{
		ID: uuid.New().String(), SessionID: sessionID,
		Role: "user", Content: content, Source: in.Source,
		CreatedAt: now,
	}
	assistantMsg := types.ChatMessage{
		ID: uuid.New().String(), SessionID: sessionID,
		Role: "assistant", Content: "", Status: "pending",
		Source: in.Source, CreatedAt: now.Add(time.Millisecond),
	}

	if _, err := w.messageStore.Create(ctx, []types.ChatMessage{userMsg, assistantMsg}); err != nil {
		return Output{}, err
	}

	execCtx, release := w.cancellations.Begin(context.Background(), sessionID)
	go func() {
		defer release()
		_ = w.pipeline.Execute(execCtx, pipeline.Input{
			Message:        assistantMsg,
			SessionID:      sessionID,
			Content:        content,
			Artifacts:      artifacts,
			ModelConfig:    in.ModelConfig,
			ResponseTo:     in.ResponseTo,
			Source:         in.Source,
			DisableHistory: in.DisableHistory,
			ErrorPrefix:    in.ErrorPrefix,
			Timeout:        in.Timeout,
			Finally:        in.Finally,
		})
	}()

	return Output{UserMessage: userMsg, AssistantMessage: assistantMsg}, nil
}
