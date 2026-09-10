package ocr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
)

const recognizePath = "/v1/recognize"

type Client struct {
	http    *http.Client
	baseURL string
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		http:    &http.Client{Timeout: timeout},
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

type recognizeResponse struct {
	Text string `json:"text"`
}

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// Recognize posts rendered pages and returns their text.
//
// The status code carries the retry decision: the service answers 4xx for
// something that will fail identically next time, so those are wrapped as
// unsupported rather than returned as errors the job would retry forever.
func (c *Client) Recognize(ctx context.Context, pages [][]byte, languages []string) (string, error) {
	body, contentType, err := form(pages, languages)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+recognizePath, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("ocr request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		var failure errorResponse
		_ = json.NewDecoder(resp.Body).Decode(&failure)

		if resp.StatusCode < http.StatusInternalServerError {
			return "", fmt.Errorf("%w: ocr rejected the document: %s",
				constants.ErrUnsupportedMedia, failure.Error)
		}
		return "", fmt.Errorf("ocr failed: %s (%d)", failure.Error, resp.StatusCode)
	}

	var decoded recognizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("ocr response: %w", err)
	}
	return decoded.Text, nil
}

func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/healthz", nil)
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("ocr health: status %d", resp.StatusCode)
	}
	return nil
}

func form(pages [][]byte, languages []string) (*bytes.Buffer, string, error) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	for index, page := range pages {
		part, err := writer.CreateFormFile("pages", fmt.Sprintf("page-%d.png", index))
		if err != nil {
			return nil, "", err
		}
		if _, err := part.Write(page); err != nil {
			return nil, "", err
		}
	}

	if len(languages) > 0 {
		if err := writer.WriteField("languages", strings.Join(languages, ",")); err != nil {
			return nil, "", err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	return &buffer, writer.FormDataContentType(), nil
}
