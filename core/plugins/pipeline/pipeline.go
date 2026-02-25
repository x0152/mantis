package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"mantis/core/agents"
	modelplugin "mantis/core/plugins/model"
	"mantis/core/protocols"
	"mantis/core/types"
	"mantis/shared"
)

type Input struct {
	Message        types.ChatMessage
	SessionID      string
	Content        string
	Artifacts      *shared.ArtifactStore
	ModelConfig    modelplugin.Input
	ResponseTo     protocols.ResponseTo
	Source         string
	DisableHistory bool
	ErrorPrefix    string
	Timeout        time.Duration
	Finally        func()
}

type Result struct {
	Message  types.ChatMessage
	Outgoing []protocols.FileAttachment
	Err      error
	SendErr  error
}

type SSHStep struct {
	ToolName string
	Task     string
	Result   string
}

type MemoryExtractor interface {
	Extract(ctx context.Context, userContent, assistantContent string, sshSteps []SSHStep)
}

type RequestHandlePipeline struct {
	agent            *agents.MantisAgent
	buffer           *shared.Buffer
	messageStore     protocols.Store[string, types.ChatMessage]
	modelStore       protocols.Store[string, types.Model]
	modelResolver    *modelplugin.Resolver
	memoryExtractor  MemoryExtractor
}

func New(
	agent *agents.MantisAgent,
	buffer *shared.Buffer,
	messageStore protocols.Store[string, types.ChatMessage],
	modelStore protocols.Store[string, types.Model],
	modelResolver *modelplugin.Resolver,
	memoryExtractor MemoryExtractor,
) *RequestHandlePipeline {
	return &RequestHandlePipeline{
		agent:           agent,
		buffer:          buffer,
		messageStore:    messageStore,
		modelStore:      modelStore,
		modelResolver:   modelResolver,
		memoryExtractor: memoryExtractor,
	}
}

func (p *RequestHandlePipeline) Execute(ctx context.Context, in Input) Result {
	if in.Finally != nil {
		defer in.Finally()
	}
	if in.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, in.Timeout)
		defer cancel()
	}

	modelOut, err := p.modelResolver.Execute(ctx, in.ModelConfig)
	if err != nil {
		return p.fail(ctx, in, err)
	}
	modelID := strings.TrimSpace(modelOut.ModelID)
	if modelID == "" {
		return p.fail(ctx, in, fmt.Errorf("model not configured"))
	}

	if model, err := shared.ResolveModel(ctx, p.modelStore, modelID); err == nil {
		in.Message.ModelName = model.Name
	}

	var replyChannel, replyTo string
	if in.ResponseTo != nil {
		replyChannel = in.ResponseTo.Channel()
		replyTo = in.ResponseTo.Recipient()
	}

	stream, runErr := p.agent.Execute(ctx, agents.MantisInput{
		SessionID:      in.SessionID,
		ModelID:        modelID,
		Content:        in.Content,
		Artifacts:      in.Artifacts,
		RequestID:      in.Message.ID,
		Source:         in.Source,
		ReplyChannel:   replyChannel,
		ReplyTo:        replyTo,
		DisableHistory: in.DisableHistory,
	})

	var content string
	var steps []types.Step
	if runErr == nil && stream != nil {
		content, steps, runErr = p.collectStream(in.Message.ID, stream)
	}

	msg := p.finalizeMessage(in.Message, content, steps, runErr, in.ErrorPrefix)
	p.saveMessage(msg)

	if p.memoryExtractor != nil && msg.Status != "error" && in.Content != "" && msg.Content != "" {
		sshSteps := collectSSHSteps(steps)
		go p.memoryExtractor.Extract(context.Background(), in.Content, msg.Content, sshSteps)
	}

	outgoing := p.collectOutgoing(in.Message.ID, in.Artifacts)
	sendErr := p.send(ctx, in.ResponseTo, msg.Content, steps, outgoing)

	p.cleanBuffer(in.Message.ID)

	return Result{Message: msg, Outgoing: outgoing, Err: runErr, SendErr: sendErr}
}

func (p *RequestHandlePipeline) fail(ctx context.Context, in Input, err error) Result {
	msg := p.finalizeMessage(in.Message, "", nil, err, in.ErrorPrefix)
	p.saveMessage(msg)
	sendErr := p.send(ctx, in.ResponseTo, msg.Content, nil, nil)
	p.cleanBuffer(in.Message.ID)
	return Result{Message: msg, Err: err, SendErr: sendErr}
}

func (p *RequestHandlePipeline) collectStream(requestID string, stream <-chan types.StreamEvent) (string, []types.Step, error) {
	var sb strings.Builder
	var steps []types.Step
	var err error
	stepIdx := map[string]int{}

	for event := range stream {
		switch event.Type {
		case "error":
			err = errors.New(event.Delta)
		case "text":
			sb.WriteString(event.Delta)
			if p.buffer != nil {
				p.buffer.SetContent(requestID, sb.String())
			}
		case "tool_start":
			var step types.Step
			_ = json.Unmarshal([]byte(event.Delta), &step)
			step.ContentOffset = sb.Len()
			stepIdx[step.ID] = len(steps)
			steps = append(steps, step)
			if p.buffer != nil {
				p.buffer.SetStep(requestID, step)
			}
		case "tool_meta":
			if idx, ok := stepIdx[event.ToolID]; ok {
				steps[idx].LogID = event.LogID
				steps[idx].ModelName = event.ModelName
				if p.buffer != nil {
					p.buffer.SetStep(requestID, steps[idx])
				}
			}
		case "tool_end":
			if idx, ok := stepIdx[event.ToolID]; ok {
				steps[idx].Status = "completed"
				steps[idx].Result = event.Delta
				steps[idx].LogID = event.LogID
				steps[idx].ModelName = event.ModelName
				steps[idx].FinishedAt = time.Now().UTC().Format(time.RFC3339)
				if p.buffer != nil {
					p.buffer.SetStep(requestID, steps[idx])
				}
			}
		}
	}

	return sb.String(), steps, err
}

func (p *RequestHandlePipeline) finalizeMessage(msg types.ChatMessage, content string, steps []types.Step, err error, errorPrefix string) types.ChatMessage {
	msg.Content = content
	if len(steps) > 0 {
		if data, e := json.Marshal(steps); e == nil {
			msg.Steps = data
		}
	}
	if err == nil {
		msg.Status = ""
		return msg
	}
	prefix := strings.TrimSpace(errorPrefix)
	if prefix == "" {
		prefix = "[Error]"
	}
	errText := fmt.Sprintf("%s %v", prefix, err)
	if msg.Content != "" {
		msg.Content += "\n\n" + errText
	} else {
		msg.Content = errText
	}
	msg.Status = "error"
	return msg
}

func (p *RequestHandlePipeline) saveMessage(msg types.ChatMessage) {
	if _, err := p.messageStore.Update(context.Background(), []types.ChatMessage{msg}); err != nil {
		log.Printf("pipeline: save message: %v", err)
	}
}

func (p *RequestHandlePipeline) collectOutgoing(requestID string, artifacts *shared.ArtifactStore) []protocols.FileAttachment {
	if artifacts == nil {
		return nil
	}
	queued := artifacts.TakeOutgoing(requestID)
	if len(queued) == 0 {
		return nil
	}
	var files []protocols.FileAttachment
	for _, q := range queued {
		a, ok := artifacts.Get(q.ArtifactID)
		if !ok || len(a.Bytes) == 0 {
			continue
		}
		name := q.FileName
		if name == "" {
			name = a.Name
		}
		if name == "" {
			name = "attachment"
		}
		files = append(files, protocols.FileAttachment{
			FileName: name,
			MimeType: a.MIME,
			Data:     a.Bytes,
			Caption:  q.Caption,
		})
	}
	return files
}

func (p *RequestHandlePipeline) send(ctx context.Context, sender protocols.ResponseTo, text string, steps []types.Step, files []protocols.FileAttachment) error {
	if sender == nil {
		return nil
	}
	err := sender.Execute(ctx, protocols.DeliveryRequest{
		Text:  text,
		Steps: steps,
		Files: files,
	})
	if err != nil {
		log.Printf("pipeline: send: %v", err)
	}
	return err
}

func collectSSHSteps(steps []types.Step) []SSHStep {
	var out []SSHStep
	for _, s := range steps {
		if !strings.HasPrefix(s.Tool, "ssh_") || strings.HasPrefix(s.Tool, "ssh_download_") || strings.HasPrefix(s.Tool, "ssh_upload_") {
			continue
		}
		if s.Status != "completed" || s.Result == "" {
			continue
		}
		var args struct {
			Task string `json:"task"`
		}
		_ = json.Unmarshal([]byte(s.Args), &args)
		out = append(out, SSHStep{ToolName: s.Tool, Task: args.Task, Result: s.Result})
	}
	return out
}

func (p *RequestHandlePipeline) cleanBuffer(requestID string) {
	if p.buffer != nil {
		p.buffer.Delete(requestID)
	}
}
