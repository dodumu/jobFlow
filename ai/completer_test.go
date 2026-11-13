package ai

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestFakeDefaultReply(t *testing.T) {
	resp, err := Fake{}.Complete(context.Background(), Request{Prompt: "anything"})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if resp.Text == "" || resp.Model != "fake-model" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestFakeCustomReplyAndError(t *testing.T) {
	upper := Fake{Reply: func(req Request) (string, error) {
		return strings.ToUpper(req.Prompt), nil
	}}

	resp, err := upper.Complete(context.Background(), Request{Prompt: "hello"})
	if err != nil || resp.Text != "HELLO" {
		t.Errorf("got %q, %v; want HELLO, nil", resp.Text, err)
	}

	boom := errors.New("provider down")
	failing := Fake{Reply: func(req Request) (string, error) { return "", boom }}

	if _, err := failing.Complete(context.Background(), Request{}); !errors.Is(err, boom) {
		t.Errorf("error = %v, want %v", err, boom)
	}
}

func TestFakeRespectsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := (Fake{}).Complete(ctx, Request{}); !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}
