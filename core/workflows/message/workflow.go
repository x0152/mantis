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
	pipeline     *pipeline.RequestHandlePipeline
	messageStore protocols.Store[string, types.ChatMessage]
	artifactMgr  *artifactplugin.Manager
}

func New(
	messageStore protocols.Store[string, types.ChatMessage],
	modelStore protocols.Store[string, types.Model],
	agent *agents.MantisAgent,
	buffer *shared.Buffer,
	modelResolver *modelplugin.Resolver,
	artifactMgr *artifactplugin.Manager,
) *Workflow {
	if artifactMgr == nil {
		artifactMgr = artifactplugin.NewManager(nil)
	}
	return &Workflow{
		pipeline:     pipeline.New(agent, buffer, messageStore, modelStore, modelResolver),
		messageStore: messageStore,
		artifactMgr:  artifactMgr,
	}
}

func (w *Workflow) Execute(ctx context.Context, in Input) (Output, error) {
	sessionID := strings.TrimSpace(in.SessionID)
	if sessionID == "" {
		return Output{}, fmt.Errorf("session_id is required")
	}

	artifacts := w.artifactMgr.ForSession(sessionID)
	for _, f := range in.Incoming {
		if len(f.Data) == 0 {
			continue
		}
		name := f.FileName
		if name == "" {
			name = "attachment"
		}
		artifacts.Put(name, f.Data, f.MimeType)
	}

	now := time.Now()
	userMsg := types.ChatMessage{
		ID: uuid.New().String(), SessionID: sessionID,
		Role: "user", Content: in.Content, Source: in.Source,
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

	go func() {
		_ = w.pipeline.Execute(context.Background(), pipeline.Input{
			Message:        assistantMsg,
			SessionID:      sessionID,
			Content:        in.Content,
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
