package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const okBody = `{
	"model": "test-model",
	"content": [{"type": "text", "text": "Hello "}, {"type": "text", "text": "world"}],
	"usage": {"input_tokens": 12, "output_tokens": 5}
}`

// newTestClient points a client at a fake provider.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *atomic.Int32) {
	t.Helper()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Config{
		APIKey:       "secret-test-key",
		BaseURL:      server.URL,
		Model:        "test-model",
		MaxRetries:   2,
		RetryBackoff: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	return client, &calls
}

func TestCompleteSuccess(t *testing.T) {
	client, calls := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "secret-test-key" {
			t.Errorf("x-api-key header = %q", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") == "" {
			t.Error("anthropic-version header is missing")
		}
		if r.URL.Path != "/v1/messages" {
			t.Errorf("path = %q", r.URL.Path)
		}

		var sent apiRequest
		if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
			t.Errorf("decoding request: %v", err)
		}
		if sent.Model != "test-model" || sent.System != "be brief" ||
			len(sent.Messages) != 1 || sent.Messages[0].Content != "say hello" {
			t.Errorf("unexpected request body: %+v", sent)
		}

		w.Write([]byte(okBody))
	})

	resp, err := client.Complete(context.Background(), Request{System: "be brief", Prompt: "say hello"})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}

	if resp.Text != "Hello world" {
		t.Errorf("text = %q, want %q", resp.Text, "Hello world")
	}
	if resp.Model != "test-model" || resp.InputTokens != 12 || resp.OutputTokens != 5 {
		t.Errorf("unexpected response: %+v", resp)
	}
	if calls.Load() != 1 {
		t.Errorf("calls = %d, want 1", calls.Load())
	}
}

func TestCompleteRetriesTemporaryFailures(t *testing.T) {
	var attempts atomic.Int32

	client, calls := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) <= 2 {
			w.WriteHeader(529)
			w.Write([]byte(`{"error": {"type": "overloaded_error", "message": "busy"}}`))
			return
		}
		w.Write([]byte(okBody))
	})

	resp, err := client.Complete(context.Background(), Request{Prompt: "hi"})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if resp.Text != "Hello world" {
		t.Errorf("text = %q", resp.Text)
	}
	if calls.Load() != 3 {
		t.Errorf("calls = %d, want 3 (two failures, then success)", calls.Load())
	}
}

func TestCompleteDoesNotRetryBadRequests(t *testing.T) {
	client, calls := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": {"type": "invalid_request_error", "message": "max_tokens too large"}}`))
	})

	_, err := client.Complete(context.Background(), Request{Prompt: "hi"})

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want an *APIError", err)
	}
	if apiErr.StatusCode != 400 || apiErr.Message != "max_tokens too large" {
		t.Errorf("unexpected APIError: %+v", apiErr)
	}
	if calls.Load() != 1 {
		t.Errorf("calls = %d, want 1 because a 400 must not be retried", calls.Load())
	}
}

func TestCompleteGivesUpAfterMaxRetries(t *testing.T) {
	client, calls := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := client.Complete(context.Background(), Request{Prompt: "hi"})
	if err == nil {
		t.Fatal("expected an error after repeated failures")
	}
	if calls.Load() != 3 { // 1 attempt + 2 retries
		t.Errorf("calls = %d, want 3", calls.Load())
	}
}

func TestCompleteStopsWhenContextIsCancelled(t *testing.T) {
	client, calls := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(okBody))
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Complete(ctx, Request{Prompt: "hi"})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
	if calls.Load() != 0 {
		t.Errorf("calls = %d, want 0", calls.Load())
	}
}

func TestErrorsNeverContainTheAPIKey(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"type": "authentication_error", "message": "invalid x-api-key"}}`))
	})

	_, err := client.Complete(context.Background(), Request{Prompt: "hi"})
	if err == nil {
		t.Fatal("expected an error")
	}
	if strings.Contains(err.Error(), "secret-test-key") {
		t.Errorf("the API key leaked into the error: %v", err)
	}
}

func TestNewFromEnvRequiresAKey(t *testing.T) {
	t.Setenv("AI_API_KEY", "")

	_, err := NewFromEnv()
	if !errors.Is(err, ErrNoAPIKey) {
		t.Errorf("error = %v, want ErrNoAPIKey", err)
	}
}
