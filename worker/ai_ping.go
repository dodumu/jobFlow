package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"jobFlow/ai"
	"jobFlow/models"
)

func NewAIPingHandler(client ai.Completer) Handler {
	return func(ctx context.Context, job models.AIJob) (Output, error) {
		resp, err := client.Complete(ctx, ai.Request{
			System:    "You are a connectivity check. Answer with only one work.",
			Prompt:    "Relpy with the single word: pong",
			MaxTokens: 20,
		})
		if err != nil {
			return Output{}, err
		}

		result, err := json.Marshal(map[string]any{
			"reply":         resp.Text,
			"input_tokens":  resp.InputTokens,
			"output_tokens": resp.OutputTokens,
		})

		if err != nil {
			return Output{}, fmt.Errorf("encoding ping resutl: %w", err)
		}

		return Output{
			Result:        string(result),
			Model:         resp.Model,
			PromptVersion: "ping-v1",
		}, nil
	}
}
