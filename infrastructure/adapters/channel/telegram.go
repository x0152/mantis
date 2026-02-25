package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type FileAttachment struct {
	FileName string
	MimeType string
	Data     []byte
	Caption  string
}

type Reply struct {
	Text        string
	Files       []FileAttachment
	ReplyMarkup json.RawMessage // optional reply_markup for sendMessage
}

type MessageHandler func(ctx context.Context, chatID string, text string, files []FileAttachment) (Reply, error)

type Telegram struct {
	token   string
	allowed map[int64]bool
	handler MessageHandler
	client  *http.Client
}

func NewTelegram(token string, allowedIDs []int64, handler MessageHandler) *Telegram {
	m := make(map[int64]bool, len(allowedIDs))
	for _, id := range allowedIDs {
		m[id] = true
	}
	return &Telegram{
		token:   token,
		allowed: m,
		handler: handler,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (t *Telegram) Execute(ctx context.Context) error {
	// Best-effort: register bot commands so they show up in Telegram UI.
	if err := t.setMyCommands(ctx); err != nil {
		log.Printf("telegram: setMyCommands: %v", err)
	}

	offset := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		updates, err := t.getUpdates(ctx, offset)
		if err != nil {
			log.Printf("telegram: getUpdates: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, u := range updates {
			offset = u.UpdateID + 1
			switch {
			case u.CallbackQuery != nil:
				if len(t.allowed) > 0 && !t.allowed[u.CallbackQuery.From.ID] {
					continue
				}
				go t.handleCallback(ctx, u.CallbackQuery)
			case u.Message != nil:
				if u.Message.Text == "" && u.Message.Caption == "" && u.Message.Document == nil && u.Message.Audio == nil && u.Message.Voice == nil && len(u.Message.Photo) == 0 {
					continue
				}
				if len(t.allowed) > 0 && !t.allowed[u.Message.From.ID] {
					continue
				}
				go t.handle(ctx, u.Message)
			default:
				continue
			}
		}
	}
}

func (t *Telegram) setMyCommands(ctx context.Context) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setMyCommands", t.token)
	payload, _ := json.Marshal(map[string]any{
		"commands": []map[string]string{
			{"command": "start", "description": "Start / show welcome message"},
			{"command": "model", "description": "Switch model"},
			{"command": "reset", "description": "Reset chat context"},
			{"command": "voice", "description": "Read last message aloud"},
		},
	})
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram API error %d: %s", resp.StatusCode, string(body))
	}
	var result struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}
	if !result.OK {
		return fmt.Errorf("telegram API: %s", string(body))
	}
	return nil
}

func (t *Telegram) handleCallback(ctx context.Context, cq *tgCallbackQuery) {
	if cq == nil || cq.Message == nil {
		return
	}

	chatID := cq.Message.Chat.ID
	data := strings.TrimSpace(cq.Data)

	// Best-effort: stop the loading spinner.
	_ = t.answerCallbackQuery(ctx, cq.ID)

	if data == "" {
		return
	}

	// Route callback data to text commands so the app layer can reuse existing command logic.
	var text string
	if strings.HasPrefix(data, "model:") {
		id := strings.TrimSpace(strings.TrimPrefix(data, "model:"))
		if id != "" {
			text = "/model " + id
		}
	}
	if text == "" {
		return
	}

	reply, err := t.handler(ctx, fmt.Sprintf("%d", chatID), text, nil)
	if err != nil {
		reply = Reply{Text: fmt.Sprintf("Error: %v", err)}
	}

	for _, f := range reply.Files {
		if len(f.Data) == 0 {
			continue
		}
		if err := t.sendDocument(ctx, chatID, f); err != nil {
			log.Printf("telegram: sendDocument: %v", err)
		}
	}

	outText := reply.Text
	if outText == "" && len(reply.Files) == 0 {
		outText = "(empty response)"
	}
	if outText == "" {
		return
	}

	markup := reply.ReplyMarkup
	for i, chunk := range splitMessage(outText, 4096) {
		rm := json.RawMessage(nil)
		if i == 0 {
			rm = markup
		}
		if err := t.sendMessageWithMarkup(ctx, chatID, chunk, rm); err != nil {
			log.Printf("telegram: sendMessage: %v", err)
		}
	}
}

func (t *Telegram) handle(ctx context.Context, msg *tgMessage) {
	done := make(chan struct{})
	go t.typingLoop(ctx, msg.Chat.ID, done)
	defer close(done)

	text := msg.Text
	if text == "" {
		text = msg.Caption
	}

	const maxBytes = 10 * 1024 * 1024
	var incoming []FileAttachment
	appendFile := func(fileID string, fileSize int64, name, mime string) error {
		if fileID == "" {
			return nil
		}
		if fileSize > 0 && fileSize > maxBytes {
			return fmt.Errorf("file too large (%d bytes, max %d)", fileSize, maxBytes)
		}
		filePath, err := t.getFilePath(ctx, fileID)
		if err != nil {
			return err
		}
		data, err := t.downloadFileByPath(ctx, filePath, maxBytes)
		if err != nil {
			return err
		}
		if name == "" {
			name = path.Base(filePath)
		}
		if name == "" {
			name = "file"
		}
		incoming = append(incoming, FileAttachment{
			FileName: name,
			MimeType: mime,
			Data:     data,
		})
		return nil
	}

	// document
	if msg.Document != nil {
		name := msg.Document.FileName
		if name == "" {
			name = "document"
		}
		if err := appendFile(msg.Document.FileID, msg.Document.FileSize, name, msg.Document.MimeType); err != nil {
			_ = t.sendMessage(ctx, msg.Chat.ID, "Error: "+err.Error())
			return
		}
	}

	// audio
	if msg.Audio != nil {
		name := msg.Audio.FileName
		if name == "" {
			name = "audio"
		}
		if err := appendFile(msg.Audio.FileID, msg.Audio.FileSize, name, msg.Audio.MimeType); err != nil {
			_ = t.sendMessage(ctx, msg.Chat.ID, "Error: "+err.Error())
			return
		}
	}

	// voice (often OGG/OPUS)
	if msg.Voice != nil {
		if err := appendFile(msg.Voice.FileID, msg.Voice.FileSize, "voice.ogg", msg.Voice.MimeType); err != nil {
			_ = t.sendMessage(ctx, msg.Chat.ID, "Error: "+err.Error())
			return
		}
	}

	if len(msg.Photo) > 0 {
		best := msg.Photo[len(msg.Photo)-1]
		photoName := fmt.Sprintf("photo_%s.jpg", time.Now().UTC().Format("20060102_150405"))
		if err := appendFile(best.FileID, best.FileSize, photoName, "image/jpeg"); err != nil {
			_ = t.sendMessage(ctx, msg.Chat.ID, "Error: "+err.Error())
			return
		}
	}

	if text == "" && len(incoming) > 0 {
		text = "User attached file(s)."
	}

	reply, err := t.handler(ctx, fmt.Sprintf("%d", msg.Chat.ID), text, incoming)

	if err != nil {
		reply = Reply{Text: fmt.Sprintf("Error: %v", err)}
	}

	// If there are attachments, send them first. This avoids "sent" text arriving
	// minutes before the actual file finishes uploading.
	for _, f := range reply.Files {
		if len(f.Data) == 0 {
			continue
		}
		if err := t.sendDocument(ctx, msg.Chat.ID, f); err != nil {
			log.Printf("telegram: sendDocument: %v", err)
		}
	}

	if reply.Text != "" {
		markup := reply.ReplyMarkup
		for i, chunk := range splitMessage(reply.Text, 4096) {
			rm := json.RawMessage(nil)
			if i == 0 {
				rm = markup
			}
			if err := t.sendMessageWithMarkup(ctx, msg.Chat.ID, chunk, rm); err != nil {
				log.Printf("telegram: sendMessage: %v", err)
			}
		}
	}
}

func (t *Telegram) typingLoop(ctx context.Context, chatID int64, done <-chan struct{}) {
	t.sendChatAction(ctx, chatID, "typing")
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.sendChatAction(ctx, chatID, "typing")
		}
	}
}

func (t *Telegram) sendChatAction(ctx context.Context, chatID int64, action string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendChatAction", t.token)
	payload, _ := json.Marshal(map[string]any{"chat_id": chatID, "action": action})
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(payload)))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

type tgUpdate struct {
	UpdateID       int              `json:"update_id"`
	Message        *tgMessage       `json:"message"`
	CallbackQuery  *tgCallbackQuery `json:"callback_query"`
	EditedMessage  *tgMessage       `json:"edited_message"`
	ChannelPost    *tgMessage       `json:"channel_post"`
	EditedPost     *tgMessage       `json:"edited_channel_post"`
	BusinessMsg    *tgMessage       `json:"business_message"`
	EditedBusiness *tgMessage       `json:"edited_business_message"`
}

type tgCallbackQuery struct {
	ID      string     `json:"id"`
	From    tgFrom     `json:"from"`
	Message *tgMessage `json:"message"`
	Data    string     `json:"data"`
}

type tgFrom struct {
	ID int64 `json:"id"`
}

type tgMessage struct {
	Chat struct {
		ID int64 `json:"id"`
	} `json:"chat"`
	From struct {
		ID int64 `json:"id"`
	} `json:"from"`
	Text     string      `json:"text"`
	Caption  string      `json:"caption"`
	Document *tgDocument `json:"document"`
	Audio    *tgAudio    `json:"audio"`
	Voice    *tgVoice    `json:"voice"`
	Photo    []tgPhoto   `json:"photo"`
}

type tgDocument struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	FileSize int64  `json:"file_size"`
}

type tgAudio struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	FileSize int64  `json:"file_size"`
}

type tgVoice struct {
	FileID   string `json:"file_id"`
	MimeType string `json:"mime_type"`
	FileSize int64  `json:"file_size"`
}

type tgPhoto struct {
	FileID   string `json:"file_id"`
	FileSize int64  `json:"file_size"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

func (t *Telegram) getUpdates(ctx context.Context, offset int) ([]tgUpdate, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=30", t.token, offset)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		OK     bool       `json:"ok"`
		Result []tgUpdate `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if !result.OK {
		return nil, fmt.Errorf("telegram API: %s", string(body))
	}
	return result.Result, nil
}

func (t *Telegram) answerCallbackQuery(ctx context.Context, callbackQueryID string) error {
	if callbackQueryID == "" {
		return nil
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", t.token)
	payload, _ := json.Marshal(map[string]any{
		"callback_query_id": callbackQueryID,
	})
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram API error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (t *Telegram) downloadByFileID(ctx context.Context, fileID string, maxBytes int64) ([]byte, error) {
	filePath, err := t.getFilePath(ctx, fileID)
	if err != nil {
		return nil, err
	}
	return t.downloadFileByPath(ctx, filePath, maxBytes)
}

func (t *Telegram) getFilePath(ctx context.Context, fileID string) (string, error) {
	if fileID == "" {
		return "", fmt.Errorf("file_id is required")
	}
	u := fmt.Sprintf("https://api.telegram.org/bot%s/getFile?file_id=%s", t.token, url.QueryEscape(fileID))
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return "", err
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("telegram API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			FilePath string `json:"file_path"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if !result.OK || result.Result.FilePath == "" {
		return "", fmt.Errorf("telegram getFile failed: %s", string(body))
	}
	return result.Result.FilePath, nil
}

func (t *Telegram) downloadFileByPath(ctx context.Context, filePath string, maxBytes int64) ([]byte, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file_path is empty")
	}
	u := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", t.token, filePath)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("file too large: %d bytes (max %d)", len(data), maxBytes)
	}
	return data, nil
}

func (t *Telegram) sendMessage(ctx context.Context, chatID int64, text string) error {
	return t.sendMessageWithMarkup(ctx, chatID, text, nil)
}

func (t *Telegram) sendMessageWithMarkup(ctx context.Context, chatID int64, text string, replyMarkup json.RawMessage) error {
	// Try Telegram MarkdownV2 first. If Telegram rejects it (e.g. "can't parse entities"),
	// fall back to legacy Markdown, then plain text so the user still gets the message.
	normalized := normalizeTelegramMarkdownV2(text)
	if err := t.sendMessageRaw(ctx, chatID, normalized, "MarkdownV2", replyMarkup); err != nil {
		if errMD := t.sendMessageRaw(ctx, chatID, normalized, "Markdown", replyMarkup); errMD != nil {
			if errPlain := t.sendMessageRaw(ctx, chatID, normalized, "", replyMarkup); errPlain != nil {
				return fmt.Errorf("send markdownv2 failed: %w; markdown fallback failed: %w; plain fallback failed: %w", err, errMD, errPlain)
			}
			log.Printf("telegram: markdown rejected, sent plain: mdv2=%v md=%v", err, errMD)
			return nil
		}
		log.Printf("telegram: markdownv2 rejected, sent legacy markdown: %v", err)
	}
	return nil
}

func normalizeTelegramMarkdownV2(text string) string {
	// Telegram MarkdownV2 is not the same as "normal" Markdown.
	// This does a few safe, minimal conversions so that common LLM output renders
	// correctly instead of constantly triggering a parse error and falling back.
	if text == "" {
		return text
	}
	out := strings.ReplaceAll(text, "\r\n", "\n")
	// CommonMark/GFM → Telegram MarkdownV2.
	out = strings.ReplaceAll(out, "~~", "~") // strike
	out = strings.ReplaceAll(out, "**", "*") // bold

	lines := strings.Split(out, "\n")
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		prefix := line[:len(line)-len(trimmed)]

		// Headings (#, ##, ...) are not supported in MarkdownV2 and '#' is reserved.
		j := 0
		for j < len(trimmed) && trimmed[j] == '#' {
			j++
		}
		if j > 0 && j <= 6 && j < len(trimmed) && trimmed[j] == ' ' {
			lines[i] = prefix + strings.TrimSpace(trimmed[j:])
			continue
		}

		// Blockquotes: avoid raw '>' (reserved) and render as a plain quote marker.
		if strings.HasPrefix(trimmed, "> ") {
			lines[i] = prefix + "│ " + strings.TrimSpace(trimmed[2:])
			continue
		}

		// Lists aren't a real thing in Telegram MarkdownV2; use bullet characters to avoid '-' issues.
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			lines[i] = prefix + "• " + strings.TrimSpace(trimmed[2:])
			continue
		}

		// Ordered list like "1. item" → "1) item" ('.' can cause parse issues in MarkdownV2).
		n := 0
		for n < len(trimmed) && trimmed[n] >= '0' && trimmed[n] <= '9' {
			n++
		}
		if n > 0 && n+1 < len(trimmed) && trimmed[n] == '.' && trimmed[n+1] == ' ' {
			rest := strings.TrimSpace(trimmed[n+2:])
			lines[i] = prefix + trimmed[:n] + ") " + rest
			continue
		}
	}
	return strings.Join(lines, "\n")
}

func (t *Telegram) sendMessageRaw(ctx context.Context, chatID int64, text string, parseMode string, replyMarkup json.RawMessage) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)
	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}
	if parseMode != "" {
		payload["parse_mode"] = parseMode
	}
	if len(replyMarkup) > 0 {
		payload["reply_markup"] = replyMarkup
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram API error %d: %s", resp.StatusCode, string(respBody))
	}
	var result struct {
		OK bool `json:"ok"`
	}
	_ = json.Unmarshal(respBody, &result)
	if !result.OK {
		return fmt.Errorf("telegram API: %s", string(respBody))
	}
	return nil
}

func (t *Telegram) sendDocument(ctx context.Context, chatID int64, f FileAttachment) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", t.token)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("chat_id", fmt.Sprintf("%d", chatID))
	if f.Caption != "" {
		_ = w.WriteField("caption", f.Caption)
	}

	part, err := w.CreateFormFile("document", f.FileName)
	if err != nil {
		return err
	}
	if _, err := part.Write(f.Data); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		OK bool `json:"ok"`
	}
	_ = json.Unmarshal(body, &result)
	if !result.OK {
		return fmt.Errorf("telegram API: %s", string(body))
	}
	return nil
}

func (t *Telegram) SendVoice(ctx context.Context, chatID int64, audio []byte) error {
	u := fmt.Sprintf("https://api.telegram.org/bot%s/sendVoice", t.token)
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("chat_id", fmt.Sprintf("%d", chatID))
	part, err := w.CreateFormFile("voice", "voice.ogg")
	if err != nil {
		return err
	}
	if _, err := part.Write(audio); err != nil {
		return err
	}
	w.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", u, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram API error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (t *Telegram) SendMessage(ctx context.Context, chatID int64, text string) error {
	return t.sendMessage(ctx, chatID, text)
}

func (t *Telegram) SendDocument(ctx context.Context, chatID int64, f FileAttachment) error {
	return t.sendDocument(ctx, chatID, f)
}

func (t *Telegram) SendMessageReturningID(ctx context.Context, chatID int64, text, parseMode string) (int64, error) {
	u := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)
	payload := map[string]any{"chat_id": chatID, "text": text}
	if parseMode != "" {
		payload["parse_mode"] = parseMode
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			MessageID int64 `json:"message_id"`
		} `json:"result"`
	}
	_ = json.Unmarshal(respBody, &result)
	if !result.OK {
		return 0, fmt.Errorf("telegram API: %s", string(respBody))
	}
	return result.Result.MessageID, nil
}

func (t *Telegram) EditMessageText(ctx context.Context, chatID, messageID int64, text, parseMode string) error {
	u := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageText", t.token)
	payload := map[string]any{"chat_id": chatID, "message_id": messageID, "text": text}
	if parseMode != "" {
		payload["parse_mode"] = parseMode
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 429 {
		log.Printf("telegram: editMessageText rate limited, skipping")
		return nil
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram API error %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (t *Telegram) EditMessage(ctx context.Context, chatID, messageID int64, text string) error {
	normalized := normalizeTelegramMarkdownV2(text)
	if err := t.EditMessageText(ctx, chatID, messageID, normalized, "MarkdownV2"); err != nil {
		if err2 := t.EditMessageText(ctx, chatID, messageID, normalized, "Markdown"); err2 != nil {
			return t.EditMessageText(ctx, chatID, messageID, text, "")
		}
	}
	return nil
}

func (t *Telegram) DeleteMessage(ctx context.Context, chatID, messageID int64) error {
	u := fmt.Sprintf("https://api.telegram.org/bot%s/deleteMessage", t.token)
	payload, _ := json.Marshal(map[string]any{"chat_id": chatID, "message_id": messageID})
	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func splitMessage(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}
	var parts []string
	for len(text) > 0 {
		end := maxLen
		if end > len(text) {
			end = len(text)
		}
		parts = append(parts, text[:end])
		text = text[end:]
	}
	return parts
}
