package ai

import "context"

type Completer interface {
	Complete(ctx context.Context, req Request) (Response, error)
}

var (
	_ Completer = (*Client)(nil)
	_ Completer = Fake{}
)

type Fake struct {
	Reply func(req Request) (string, error)
}

func (f Fake) Complete(ctx context.Context, req Request) (Response, error) {
	if err := ctx.Err(); err != nil {
		return Response{}, err
	}

	text := "This is a fake AI reply"
	if f.Reply != nil {
		reply, err := f.Reply(req)
		if err != nil {
			return Response{}, err
		}
		text = reply
	}

	return Response{
		Text:         text,
		Model:        "fake-model",
		InputTokens:  len(req.Prompt) / 4,
		OutputTokens: len(text) / 4,
	}, nil
}
