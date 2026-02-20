package asr

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type GigaAM struct {
	baseURL string
	client  *http.Client
}

func NewGigaAM(baseURL string, timeout time.Duration) *GigaAM {
	return &GigaAM{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: timeout},
	}
}

func (g *GigaAM) Transcribe(ctx context.Context, audio io.Reader, format string) (string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", "audio."+format)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, audio); err != nil {
		return "", err
	}
	w.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/transcribe", &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ASR API error %d: %s", resp.StatusCode, string(body))
	}

	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "text/event-stream") {
		return g.parseSSE(resp.Body)
	}

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	return result.Text, nil
}

func (g *GigaAM) parseSSE(r io.Reader) (string, error) {
	var parts []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		raw := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		var ev struct {
			Text   string `json:"text"`
			Status string `json:"status"`
			Error  string `json:"error"`
		}
		if json.Unmarshal([]byte(raw), &ev) != nil {
			continue
		}
		if ev.Error != "" {
			return "", fmt.Errorf("ASR: %s", ev.Error)
		}
		if ev.Text != "" {
			parts = append(parts, ev.Text)
		}
		if ev.Status == "completed" {
			break
		}
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("ASR returned no transcription")
	}
	return strings.Join(parts, " "), nil
}
