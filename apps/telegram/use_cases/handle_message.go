package usecases

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	modelplugin "mantis/core/plugins/model"
	"mantis/core/protocols"
	"mantis/core/types"
	messageworkflow "mantis/core/workflows/message"
	adapter "mantis/infrastructure/adapters/channel"
	"mantis/shared"
)

type MessageInput struct {
	ChannelID string
	ChatID    string
	Text      string
	Incoming  []adapter.FileAttachment
}

type HandleMessage struct {
	sessionUC      *Session
	modelCommandUC *HandleModelCommand
	channelStore   protocols.Store[string, types.Channel]
	messageStore   protocols.Store[string, types.ChatMessage]
	workflow       *messageworkflow.Workflow
	buffer         *shared.Buffer
	asr            protocols.ASR
	tts            protocols.TTS
}

func NewHandleMessage(
	sessionUC *Session,
	modelCommandUC *HandleModelCommand,
	channelStore protocols.Store[string, types.Channel],
	messageStore protocols.Store[string, types.ChatMessage],
	workflow *messageworkflow.Workflow,
	buffer *shared.Buffer,
	asr protocols.ASR,
	tts protocols.TTS,
) *HandleMessage {
	return &HandleMessage{
		sessionUC:      sessionUC,
		modelCommandUC: modelCommandUC,
		channelStore:   channelStore,
		messageStore:   messageStore,
		workflow:       workflow,
		buffer:         buffer,
		asr:            asr,
		tts:            tts,
	}
}

func (uc *HandleMessage) Handler(channelID string) adapter.MessageHandler {
	return func(ctx context.Context, chatID string, text string, incoming []adapter.FileAttachment) (adapter.Reply, error) {
		return uc.Execute(ctx, MessageInput{
			ChannelID: channelID,
			ChatID:    chatID,
			Text:      text,
			Incoming:  incoming,
		})
	}
}

func (uc *HandleMessage) Execute(ctx context.Context, in MessageInput) (adapter.Reply, error) {
	if cmd, args := parseSlashCommandWithArgs(in.Text); cmd != "" {
		switch cmd {
		case "start":
			_, _ = uc.sessionUC.Execute(ctx, SessionModeGetOrCreate)
			return adapter.Reply{Text: "Mantis\n\nSend a message to get started.\nCommands are available via the Menu button.\n\n/model - switch model\n/reset - reset context\n/voice - read last message aloud"}, nil
		case "reset":
			if _, err := uc.sessionUC.Execute(ctx, SessionModeReset); err != nil {
				return adapter.Reply{}, err
			}
			return adapter.Reply{Text: "Context reset. Send a new message to start fresh."}, nil
		case "model":
			return uc.modelCommandUC.Execute(ctx, in.ChannelID, args)
		case "voice":
			return uc.handleVoiceCommand(ctx, in)
		}
	}

	sessionID, err := uc.sessionUC.Execute(ctx, SessionModeGetOrCreate)
	if err != nil {
		return adapter.Reply{}, err
	}
	if strings.TrimSpace(sessionID) == "" {
		return adapter.Reply{}, fmt.Errorf("failed to resolve session")
	}

	sender, err := uc.createSender(ctx, in.ChannelID, in.ChatID)
	if err != nil {
		return adapter.Reply{}, err
	}

	in.Incoming, in.Text = uc.transcribeAudio(ctx, sender, in.Incoming, in.Text)

	incomingFiles := make([]protocols.FileAttachment, 0, len(in.Incoming))
	for _, f := range in.Incoming {
		incomingFiles = append(incomingFiles, protocols.FileAttachment{
			FileName: f.FileName,
			MimeType: f.MimeType,
			Data:     f.Data,
			Caption:  f.Caption,
		})
	}

	done := make(chan struct{})
	out, err := uc.workflow.Execute(ctx, messageworkflow.Input{
		SessionID:   sessionID,
		Content:     in.Text,
		Incoming:    incomingFiles,
		Source:      "telegram",
		ResponseTo:  sender,
		ModelConfig: modelplugin.Input{ChannelID: in.ChannelID},
		Finally:     func() { close(done) },
	})
	if err != nil {
		return adapter.Reply{}, err
	}

	sender.StreamFrom(ctx, uc.buffer, out.AssistantMessage.ID, done)
	return adapter.Reply{}, nil
}

func (uc *HandleMessage) createSender(ctx context.Context, channelID, chatID string) (*adapter.TelegramResponseTo, error) {
	channels, err := uc.channelStore.Get(ctx, []string{channelID})
	if err != nil {
		return nil, err
	}
	ch, ok := channels[channelID]
	if !ok {
		return nil, fmt.Errorf("channel %q not found", channelID)
	}
	if ch.Token == "" {
		return nil, fmt.Errorf("channel %q has no token", channelID)
	}
	return adapter.NewTelegramResponseTo(ch.Token, chatID), nil
}

func (uc *HandleMessage) transcribeAudio(ctx context.Context, sender *adapter.TelegramResponseTo, files []adapter.FileAttachment, text string) ([]adapter.FileAttachment, string) {
	if uc.asr == nil {
		return files, text
	}
	var kept []adapter.FileAttachment
	for _, f := range files {
		if f.FileName != "voice.ogg" {
			kept = append(kept, f)
			continue
		}
		format := strings.TrimPrefix(f.MimeType, "audio/")
		result, err := uc.asr.Transcribe(ctx, bytes.NewReader(f.Data), format)
		if err != nil {
			log.Printf("asr: transcribe: %v", err)
			kept = append(kept, f)
			continue
		}
		result = strings.TrimSpace(result)
		if result == "" {
			continue
		}
		sender.SendQuote(ctx, "🎤 "+result)
		if text == "" || text == "User attached file(s)." {
			text = result
		} else {
			text = result + "\n\n" + text
		}
	}
	return kept, text
}

func (uc *HandleMessage) handleVoiceCommand(ctx context.Context, in MessageInput) (adapter.Reply, error) {
	if uc.tts == nil {
		return adapter.Reply{Text: "TTS is not configured."}, nil
	}

	sessionID, err := uc.sessionUC.Execute(ctx, SessionModeGetOrCreate)
	if err != nil {
		return adapter.Reply{}, err
	}

	messages, err := uc.messageStore.List(ctx, types.ListQuery{
		Filter: map[string]string{"session_id": sessionID, "role": "assistant"},
		Sort:   []types.Sort{{Field: "created_at", Dir: types.SortDirDesc}},
		Page:   types.Page{Limit: 1},
	})
	if err != nil {
		return adapter.Reply{}, err
	}
	if len(messages) == 0 {
		return adapter.Reply{Text: "No messages to read aloud."}, nil
	}

	text := strings.TrimSpace(messages[0].Content)
	if text == "" {
		return adapter.Reply{Text: "Last message is empty."}, nil
	}

	const maxTTSLen = 2000
	if len(text) > maxTTSLen {
		text = text[:maxTTSLen]
	}

	audio, err := uc.tts.Synthesize(ctx, protocols.TTSRequest{
		Text:   text,
		Format: "wav",
	})
	if err != nil {
		return adapter.Reply{Text: fmt.Sprintf("TTS error: %v", err)}, nil
	}

	sender, err := uc.createSender(ctx, in.ChannelID, in.ChatID)
	if err != nil {
		return adapter.Reply{}, err
	}
	if err := sender.SendVoice(ctx, audio); err != nil {
		return adapter.Reply{Text: fmt.Sprintf("Send error: %v", err)}, nil
	}

	return adapter.Reply{}, nil
}

func parseSlashCommandWithArgs(text string) (cmd, args string) {
	s := strings.TrimSpace(text)
	if !strings.HasPrefix(s, "/") {
		return "", ""
	}
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return "", ""
	}
	first := fields[0]

	cmd = strings.TrimPrefix(first, "/")
	if i := strings.IndexByte(cmd, '@'); i >= 0 {
		cmd = cmd[:i]
	}
	args = strings.TrimSpace(strings.TrimPrefix(s, first))
	return strings.ToLower(cmd), args
}
