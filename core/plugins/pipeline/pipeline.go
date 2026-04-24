package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mantis/core/agents"
	modelplugin "mantis/core/plugins/model"
	"mantis/core/plugins/summarizer"
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
	agent           *agents.MantisAgent
	buffer          *shared.Buffer
	messageStore    protocols.Store[string, types.ChatMessage]
	modelStore      protocols.Store[string, types.Model]
	modelResolver   *modelplugin.Resolver
	memoryExtractor MemoryExtractor
	summarizer      *summarizer.Summarizer
	attachmentDir   string
	limits          shared.Limits
}

func New(
	agent *agents.MantisAgent,
	buffer *shared.Buffer,
	messageStore protocols.Store[string, types.ChatMessage],
	modelStore protocols.Store[string, types.Model],
	modelResolver *modelplugin.Resolver,
	memoryExtractor MemoryExtractor,
	summ *summarizer.Summarizer,
	limits shared.Limits,
) *RequestHandlePipeline {
	return &RequestHandlePipeline{
		agent:           agent,
		buffer:          buffer,
		messageStore:    messageStore,
		modelStore:      modelStore,
		modelResolver:   modelResolver,
		memoryExtractor: memoryExtractor,
		summarizer:      summ,
		limits:          limits,
	}
}

func (p *RequestHandlePipeline) SetAttachmentDir(dir string) {
	p.attachmentDir = dir
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
	in.Message.ModelID = modelID
	in.Message.PresetID = strings.TrimSpace(modelOut.PresetID)
	in.Message.PresetName = strings.TrimSpace(modelOut.PresetName)
	in.Message.ModelRole = strings.TrimSpace(modelOut.ModelRole)

	if model, err := shared.ResolveModel(ctx, p.modelStore, modelID); err == nil {
		in.Message.ModelID = model.ID
		in.Message.ModelName = model.Name
	}

	if p.buffer != nil {
		p.buffer.SetSessionID(in.Message.ID, in.SessionID)
	}

	var compactStep *types.Step
	if p.summarizer != nil && !in.DisableHistory {
		if res, err := p.summarizer.MaybeCompact(ctx, summarizer.Input{
			SessionID: in.SessionID,
			RequestID: in.Message.ID,
			ModelID:   modelID,
			PresetID:  strings.TrimSpace(modelOut.PresetID),
		}); err == nil {
			compactStep = res.Step
		} else {
			log.Printf("pipeline: compact: %v", err)
		}
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

	if compactStep != nil {
		steps = append([]types.Step{*compactStep}, steps...)
	}

	stopMarker, stopped := p.classifyStop(ctx, runErr)
	if stopped {
		steps = markCancelledSteps(steps)
	}

	msg := p.finalizeMessage(in.Message, content, steps, runErr, in.ErrorPrefix, stopMarker, stopped)

	outgoing := p.collectOutgoing(in.Message.ID, in.Artifacts)

	if len(outgoing) > 0 {
		for _, f := range outgoing {
			msg.Attachments = append(msg.Attachments, types.Attachment{
				ID:       f.ArtifactID,
				FileName: f.FileName,
				MimeType: f.MimeType,
				Size:     int64(len(f.Data)),
			})
		}
		p.persistAttachments(outgoing)
	}

	p.saveMessage(msg)

	if p.memoryExtractor != nil && msg.Status != "error" && in.Content != "" && msg.Content != "" {
		sshSteps := collectSSHSteps(steps)
		go p.memoryExtractor.Extract(context.Background(), in.Content, msg.Content, sshSteps)
	}

	sendErr := p.send(ctx, in.ResponseTo, msg.Content, steps, outgoing)

	p.cleanBuffer(in.Message.ID)

	return Result{Message: msg, Outgoing: outgoing, Err: runErr, SendErr: sendErr}
}

func (p *RequestHandlePipeline) fail(ctx context.Context, in Input, err error) Result {
	stopMarker, stopped := p.classifyStop(ctx, err)
	msg := p.finalizeMessage(in.Message, "", nil, err, in.ErrorPrefix, stopMarker, stopped)
	p.saveMessage(msg)
	sendErr := p.send(ctx, in.ResponseTo, msg.Content, nil, nil)
	p.cleanBuffer(in.Message.ID)
	return Result{Message: msg, Err: err, SendErr: sendErr}
}

func (p *RequestHandlePipeline) classifyStop(ctx context.Context, runErr error) (string, bool) {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return shared.StopReasonSupervisorTimeout(p.limits.SupervisorTimeout), true
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return shared.StopReasonUser(), true
	}
	if runErr != nil && strings.Contains(runErr.Error(), "max iterations reached") {
		return shared.StopReasonSupervisorIterations(p.limits.SupervisorMaxIterations), true
	}
	return "", false
}

func (p *RequestHandlePipeline) collectStream(requestID string, stream <-chan types.StreamEvent) (string, []types.Step, error) {
	var sb strings.Builder
	var steps []types.Step
	var err error
	stepIdx := map[string]int{}
	inThinking := false

	closeThinking := func() {
		if inThinking {
			sb.WriteString("</think>\n\n")
			inThinking = false
		}
	}

	for event := range stream {
		switch event.Type {
		case "error":
			closeThinking()
			err = errors.New(event.Delta)
		case "text":
			closeThinking()
			sb.WriteString(event.Delta)
			if p.buffer != nil {
				p.buffer.SetContent(requestID, sb.String())
			}
		case "thinking":
			if !inThinking {
				sb.WriteString("<think>")
				inThinking = true
			}
			sb.WriteString(event.Delta)
			if p.buffer != nil {
				p.buffer.SetContent(requestID, sb.String())
			}
		case "tool_start":
			closeThinking()
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
				steps[idx].ModelID = event.ModelID
				steps[idx].ModelName = event.ModelName
				steps[idx].PresetID = event.PresetID
				steps[idx].PresetName = event.PresetName
				steps[idx].ModelRole = event.ModelRole
				if p.buffer != nil {
					p.buffer.SetStep(requestID, steps[idx])
				}
			}
		case "tool_end":
			if idx, ok := stepIdx[event.ToolID]; ok {
				steps[idx].Status = "completed"
				steps[idx].Result = event.Delta
				steps[idx].LogID = event.LogID
				steps[idx].ModelID = event.ModelID
				steps[idx].ModelName = event.ModelName
				steps[idx].PresetID = event.PresetID
				steps[idx].PresetName = event.PresetName
				steps[idx].ModelRole = event.ModelRole
				steps[idx].FinishedAt = time.Now().UTC().Format(time.RFC3339)
				if p.buffer != nil {
					p.buffer.SetStep(requestID, steps[idx])
				}
			}
		}
	}
	closeThinking()

	return sb.String(), steps, err
}

func (p *RequestHandlePipeline) finalizeMessage(msg types.ChatMessage, content string, steps []types.Step, err error, errorPrefix string, stopMarker string, stopped bool) types.ChatMessage {
	msg.Content = content
	if len(steps) > 0 {
		if data, e := json.Marshal(steps); e == nil {
			msg.Steps = data
		}
	}
	now := time.Now().UTC()
	msg.FinishedAt = &now
	if stopped {
		if stopMarker == "" {
			stopMarker = shared.StopReasonUser()
		}
		if msg.Content != "" {
			msg.Content += "\n\n" + stopMarker
		} else {
			msg.Content = stopMarker
		}
		msg.Status = "cancelled"
		return msg
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

func markCancelledSteps(steps []types.Step) []types.Step {
	for i := range steps {
		s := &steps[i]
		if s.Status == "completed" {
			continue
		}
		if s.Status == "running" || strings.Contains(strings.ToLower(s.Result), "context canceled") {
			s.Status = "cancelled"
			s.Result = ""
			if s.FinishedAt == "" {
				s.FinishedAt = time.Now().UTC().Format(time.RFC3339)
			}
		}
	}
	return steps
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
			ArtifactID: q.ArtifactID,
			FileName:   name,
			MimeType:   a.MIME,
			Data:       a.Bytes,
			Caption:    q.Caption,
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

func (p *RequestHandlePipeline) persistAttachments(files []protocols.FileAttachment) {
	if p.attachmentDir == "" {
		return
	}
	if err := os.MkdirAll(p.attachmentDir, 0755); err != nil {
		log.Printf("pipeline: mkdir attachments: %v", err)
		return
	}
	for _, f := range files {
		if f.ArtifactID == "" || len(f.Data) == 0 {
			continue
		}
		dataPath := filepath.Join(p.attachmentDir, f.ArtifactID)
		if err := os.WriteFile(dataPath, f.Data, 0644); err != nil {
			log.Printf("pipeline: write attachment %s: %v", f.ArtifactID, err)
			continue
		}
		meta, _ := json.Marshal(map[string]string{"mime": f.MimeType, "name": f.FileName})
		_ = os.WriteFile(dataPath+".json", meta, 0644)
	}
}

func (p *RequestHandlePipeline) cleanBuffer(requestID string) {
	if p.buffer != nil {
		p.buffer.Delete(requestID)
	}
}
