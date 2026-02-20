package channel

import (
	"context"
	"fmt"
	"html"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"mantis/core/protocols"
	"mantis/core/types"
	"mantis/shared"
)

const streamInterval = 3 * time.Second

type TelegramResponseTo struct {
	token       string
	recipient   string
	streamMsgID int64
}

func NewTelegramResponseTo(token, recipient string) *TelegramResponseTo {
	return &TelegramResponseTo{
		token:     strings.TrimSpace(token),
		recipient: strings.TrimSpace(recipient),
	}
}

func (s *TelegramResponseTo) Recipient() string { return s.recipient }
func (s *TelegramResponseTo) Channel() string   { return "telegram" }

func (s *TelegramResponseTo) Execute(ctx context.Context, req protocols.DeliveryRequest) error {
	if s == nil || s.token == "" {
		return fmt.Errorf("telegram response_to token is empty")
	}

	chatID, err := strconv.ParseInt(s.recipient, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid telegram recipient/chat id %q: %w", s.recipient, err)
	}

	tg := s.newTG()

	for _, file := range req.Files {
		if len(file.Data) == 0 {
			continue
		}
		if err := tg.SendDocument(ctx, chatID, FileAttachment{
			FileName: file.FileName,
			MimeType: file.MimeType,
			Data:     file.Data,
			Caption:  file.Caption,
		}); err != nil {
			return err
		}
	}

	text := finalText(req.Text, req.Steps)
	if text == "" {
		return nil
	}

	if msgID := atomic.LoadInt64(&s.streamMsgID); msgID > 0 {
		_ = tg.DeleteMessage(ctx, chatID, msgID)
	}
	for _, chunk := range splitMessage(text, 4096) {
		if err := tg.SendMessage(ctx, chatID, chunk); err != nil {
			return err
		}
	}
	return nil
}

func (s *TelegramResponseTo) StreamFrom(ctx context.Context, buf *shared.Buffer, bufferID string, done <-chan struct{}) {
	chatID, err := strconv.ParseInt(s.recipient, 10, 64)
	if err != nil {
		<-done
		return
	}

	tg := s.newTG()

	var tgMsgID int64
	var lastText string
	ticker := time.NewTicker(streamInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			entry, ok := buf.Get(bufferID)
			if !ok {
				continue
			}
			text := formatStreamHTML(entry)
			if text == "" || text == lastText {
				continue
			}
			lastText = text
			if tgMsgID == 0 {
				id, sendErr := tg.SendMessageReturningID(ctx, chatID, text, "HTML")
				if sendErr == nil && id > 0 {
					tgMsgID = id
					atomic.StoreInt64(&s.streamMsgID, id)
				}
			} else {
				_ = tg.EditMessageText(ctx, chatID, tgMsgID, text, "HTML")
			}
		}
	}
}

func (s *TelegramResponseTo) SendVoice(ctx context.Context, audio []byte) error {
	chatID, err := strconv.ParseInt(s.recipient, 10, 64)
	if err != nil {
		return err
	}
	return s.newTG().SendVoice(ctx, chatID, audio)
}

func (s *TelegramResponseTo) SendQuote(ctx context.Context, text string) {
	chatID, err := strconv.ParseInt(s.recipient, 10, 64)
	if err != nil {
		return
	}
	_ = s.newTG().SendMessage(ctx, chatID, text)
}

func (s *TelegramResponseTo) newTG() *Telegram {
	noop := func(context.Context, string, string, []FileAttachment) (Reply, error) {
		return Reply{}, nil
	}
	return NewTelegram(s.token, nil, noop)
}

func formatStepLine(step types.Step, useHTML bool) string {
	label := step.Label
	if useHTML {
		label = "<b>" + html.EscapeString(label) + "</b>"
	}
	return "🔧 " + label
}

func interleavedText(content string, steps []types.Step, useHTML bool) string {
	if len(steps) == 0 {
		return content
	}
	sorted := make([]types.Step, len(steps))
	copy(sorted, steps)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ContentOffset < sorted[j].ContentOffset
	})
	var sb strings.Builder
	pos := 0
	for _, step := range sorted {
		if off := step.ContentOffset; off > pos && off <= len(content) {
			chunk := strings.TrimSpace(content[pos:off])
			if chunk != "" {
				if useHTML {
					chunk = html.EscapeString(chunk)
				}
				sb.WriteString(chunk)
				sb.WriteByte('\n')
			}
			pos = off
		}
		sb.WriteString(formatStepLine(step, useHTML))
		sb.WriteByte('\n')
	}
	if pos < len(content) {
		chunk := strings.TrimSpace(content[pos:])
		if chunk != "" {
			if useHTML {
				chunk = html.EscapeString(chunk)
			}
			sb.WriteString(chunk)
		}
	}
	return strings.TrimSpace(sb.String())
}

func finalText(content string, steps []types.Step) string {
	if len(steps) == 0 {
		return strings.TrimSpace(content)
	}
	maxOffset := 0
	for _, s := range steps {
		if s.ContentOffset > maxOffset {
			maxOffset = s.ContentOffset
		}
	}
	if maxOffset >= len(content) {
		return ""
	}
	return strings.TrimSpace(content[maxOffset:])
}

func formatStreamHTML(entry shared.BufferEntry) string {
	if len(entry.Steps) == 0 && entry.Content == "" {
		return ""
	}
	if len(entry.Steps) == 0 {
		return html.EscapeString(entry.Content)
	}
	return interleavedText(entry.Content, entry.Steps, true)
}
