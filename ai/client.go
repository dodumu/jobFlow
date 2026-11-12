package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL    = "https://api.anthropic.com"
	defaultModel      = "claude-sonnet-5-5"
	defaultAPIVersion = "2026-10-21"
	defaultMaxTokens  = 1024
	maxResponseBytes  = 1 << 20
	maxRetryWait      = 30 * time.Second
)

var ErrNoAPIKey = errors.New("AI_API_KEY is not set")

type Config struct {
	APIKey       string
	BaseURL      string
	Model        string
	APIVersion   string
	Timeout      time.Duration
	MaxRetries   int
	RetryBackoff time.Duration
}

type Client struct {
	cfg  Config
	http *http.Client
}

type Request struct {
	System    string
	Prompt    string
	MaxTokens int
}

type Response struct {
	Text         string
	Model        string
	InputTokens  int
	OutputTokens int
}

type APIError struct {
	StatusCode int
	Type       string
	Message    string
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	return fmt.Sprintf("ai provider returned %d (%s): %s", e.StatusCode, e.Type, e.Message)
}

func NewClient(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, ErrNoAPIKey
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	if cfg.APIVersion == "" {
		cfg.APIVersion = defaultAPIVersion
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 60 * time.Second
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = time.Second
	}

	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}, nil
}

func NewFromEnv() (*Client, error) {
	return NewClient(Config{
		APIKey:     os.Getenv("AI_API_KEY"),
		Model:      os.Getenv("AI_MODEL"),
		BaseURL:    os.Getenv("AI_BASE_URL"),
		APIVersion: os.Getenv("AI_API_VERSION"),
		MaxRetries: 2,
	})
}

type apiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type apiRequest struct {
	Model     string       `json:"model"`
	MaxTokens int          `json:"max_tokens"`
	System    string       `json:"system,omitempty"`
	Messages  []apiMessage `json:"messages"`
}

type apiResponse struct {
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type apiErrorBody struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) Complete(ctx context.Context, req Request) (Response, error) {
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}

	body, err := json.Marshal(apiRequest{
		Model:     c.cfg.Model,
		MaxTokens: maxTokens,
		System:    req.System,
		Messages:  []apiMessage{{Role: "user", Content: req.Prompt}},
	})
	if err != nil {
		return Response{}, fmt.Errorf("building ai request: %w", err)
	}
	var lastErr error

	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		resp, err := c.once(ctx, body)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if !retryable(err) || attempt == c.cfg.MaxRetries {
			break
		}

		select {
		case <-time.After(c.waitBefore(err, attempt)):
		case <-ctx.Done():
			return Response{}, ctx.Err()
		}
	}

	return Response{}, lastErr
}

func (c *Client) once(ctx context.Context, body []byte) (Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(c.cfg.BaseURL, "/")+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("creating ai request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.cfg.APIKey)
	httpReq.Header.Set("anthropic-version", c.cfg.APIVersion)

	httpResp, err := c.http.Do(httpReq)
	if err != nil {
		// The error can mention the URL, never the key, which lives in a header.
		return Response{}, fmt.Errorf("calling ai provider: %w", err)
	}
	defer httpResp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(httpResp.Body, maxResponseBytes))
	if err != nil {
		return Response{}, fmt.Errorf("reading ai response: %w", err)
	}

	if httpResp.StatusCode >= 400 {
		return Response{}, parseAPIError(httpResp, data)
	}

	var parsed apiResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return Response{}, fmt.Errorf("decoding ai response: %w", err)
	}

	var text strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}
	if text.Len() == 0 {
		return Response{}, errors.New("ai response contained no text")
	}

	return Response{
		Text:         text.String(),
		Model:        parsed.Model,
		InputTokens:  parsed.Usage.InputTokens,
		OutputTokens: parsed.Usage.OutputTokens,
	}, nil
}

func parseAPIError(resp *http.Response, data []byte) *APIError {
	apiErr := &APIError{StatusCode: resp.StatusCode, Type: "unknown", Message: http.StatusText(resp.StatusCode)}

	var body apiErrorBody
	if json.Unmarshal(data, &body) == nil && body.Error.Message != "" {
		apiErr.Type = body.Error.Type
		apiErr.Message = body.Error.Message
	}

	if seconds, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && seconds > 0 {
		apiErr.RetryAfter = time.Duration(seconds) * time.Second
	}

	return apiErr
}

// retryable reports whether trying again could help.
func retryable(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case 429, 500, 502, 503, 504, 529:
			return true
		}
		return false // 400, 401, 403, 404...: retrying cannot fix these
	}

	// No reply at all (network trouble, per-request timeout): worth another go.
	return true
}

// waitBefore chooses how long to pause before the next attempt.
func (c *Client) waitBefore(err error, attempt int) time.Duration {
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.RetryAfter > 0 {
		return min(apiErr.RetryAfter, maxRetryWait)
	}

	wait := c.cfg.RetryBackoff << attempt // 1s, 2s, 4s...
	return min(wait, maxRetryWait)
}
