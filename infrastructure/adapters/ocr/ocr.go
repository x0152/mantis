package ocr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type API struct {
	baseURL string
	client  *http.Client
}

func NewAPI(baseURL string, timeout time.Duration) *API {
	return &API{
		baseURL: baseURL,
		client:  &http.Client{Timeout: timeout},
	}
}

func (a *API) ExtractText(ctx context.Context, image io.Reader, format string) (string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", "image."+format)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, image); err != nil {
		return "", err
	}
	w.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/ocr", &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("OCR API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	return result.Text, nil
}
