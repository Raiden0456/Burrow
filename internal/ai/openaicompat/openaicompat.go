package openaicompat

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"

	"github.com/raiden0456/gonteractive/internal/ai"
)

type Provider struct {
	client openai.Client
	model  string
}

func New(apiKey, baseURL, model string) (*Provider, error) {
	if model == "" {
		return nil, errors.New("openai: model is required")
	}
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	return &Provider{client: openai.NewClient(opts...), model: model}, nil
}

func (p *Provider) Stream(ctx context.Context, req ai.Request, onDelta func(string) error) (string, error) {
	return p.run(ctx, p.params(req), onDelta)
}

func (p *Provider) Generate(ctx context.Context, req ai.Request, schema *ai.Schema) (string, error) {
	params := p.params(req)
	params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
			JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:   schema.Name,
				Strict: openai.Bool(true),
				Schema: schema.Doc,
			},
		},
	}
	return p.run(ctx, params, func(string) error { return nil })
}

func (p *Provider) Close() error { return nil }

func (p *Provider) params(req ai.Request) openai.ChatCompletionNewParams {
	msgs := make([]openai.ChatCompletionMessageParamUnion, 0, len(req.Messages)+1)
	if req.System != "" {
		msgs = append(msgs, openai.SystemMessage(req.System))
	}
	for _, m := range req.Messages {
		if m.Role == ai.RoleAssistant {
			msgs = append(msgs, openai.AssistantMessage(m.Content))
		} else {
			msgs = append(msgs, openai.UserMessage(m.Content))
		}
	}
	return openai.ChatCompletionNewParams{Model: p.model, Messages: msgs}
}

func (p *Provider) run(ctx context.Context, params openai.ChatCompletionNewParams, onDelta func(string) error) (string, error) {
	stream := p.client.Chat.Completions.NewStreaming(ctx, params)
	defer stream.Close()

	var out strings.Builder
	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta
		if delta.Refusal != "" {
			return "", ai.ErrRefused
		}
		if delta.Content == "" {
			continue
		}
		out.WriteString(delta.Content)
		if err := onDelta(delta.Content); err != nil {
			return "", err
		}
	}
	if err := stream.Err(); err != nil {
		return "", fmt.Errorf("openai: %w", err)
	}
	return out.String(), nil
}
