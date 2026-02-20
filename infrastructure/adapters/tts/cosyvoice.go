package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mantis/core/protocols"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type CosyVoice struct {
	baseURL string
	client  *http.Client
}

func NewCosyVoice(baseURL string, timeout time.Duration) *CosyVoice {
	return &CosyVoice{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *CosyVoice) Synthesize(ctx context.Context, req protocols.TTSRequest) ([]byte, error) {
	if req.Voice != nil {
		return c.synthesizeMultipart(ctx, req)
	}
	return c.synthesizeJSON(ctx, req)
}

func (c *CosyVoice) synthesizeJSON(ctx context.Context, req protocols.TTSRequest) ([]byte, error) {
	body := map[string]any{
		"model":           "cosyvoice3",
		"input":           req.Text,
		"voice":           "default",
		"response_format": orDefault(req.Format, "wav"),
	}
	if req.Instructions != "" {
		body["instructions"] = req.Instructions
	} else if req.Emotion != "" {
		body["instructions"] = req.Emotion
	}

	payload, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/audio/speech", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	return c.doRequest(httpReq)
}

func (c *CosyVoice) synthesizeMultipart(ctx context.Context, req protocols.TTSRequest) ([]byte, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	w.WriteField("text", req.Text)
	w.WriteField("response_format", orDefault(req.Format, "wav"))

	if req.Emotion != "" {
		w.WriteField("emotion", req.Emotion)
	}
	if req.Instructions != "" {
		w.WriteField("instruct", req.Instructions)
	}

	part, err := w.CreateFormFile("voice", "voice.wav")
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, bytes.NewReader(req.Voice)); err != nil {
		return nil, err
	}
	w.Close()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/tts", &buf)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", w.FormDataContentType())

	return c.doRequest(httpReq)
}

func (c *CosyVoice) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TTS API error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
