package ai

import (
	"context"
	"errors"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Role    Role
	Content string
}

type Request struct {
	System   string
	Messages []Message
}

type Provider interface {
	Stream(ctx context.Context, req Request, onDelta func(string) error) (string, error)
	Generate(ctx context.Context, req Request, schema *Schema) (string, error)
	Close() error
}

var ErrRefused = errors.New("the model refused to answer")

func Complete(ctx context.Context, p Provider, req Request) (string, error) {
	return p.Stream(ctx, req, func(string) error { return nil })
}
