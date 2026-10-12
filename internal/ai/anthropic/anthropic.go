package anthropic

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/shared/constant"

	"github.com/raiden0456/gonteractive/internal/ai"
)

const DefaultModel = "claude-opus-5-5"

var fallbackModels = map[string]bool{
	"claude-fable-5-1":  true,
	"claude-opus-5-5":   true,
	"claude-opus-5":     true,
	"claude-sonnet-5-5": true,
}

type Provider struct {
	client    sdk.Client
	model     string
	fallbacks bool
}

func New(apiKey, baseURL, model string) *Provider {
	if model == "" {
		model = DefaultModel
	}
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	return &Provider{
		client:    sdk.NewClient(opts...),
		model:     model,
		fallbacks: baseURL == "" && fallbackModels[model],
	}
}

func (p *Provider) Stream(ctx context.Context, req ai.Request, onDelta func(string) error) (string, error) {
	return p.run(ctx, p.params(req), onDelta)
}

func (p *Provider) Generate(ctx context.Context, req ai.Request, schema *ai.Schema) (string, error) {
	params := p.params(req)
	params.OutputConfig = sdk.BetaOutputConfigParam{
		Format: sdk.BetaJSONOutputFormatParam{Schema: schema.Doc},
	}
	return p.run(ctx, params, func(string) error { return nil })
}

func (p *Provider) Close() error { return nil }

func (p *Provider) params(req ai.Request) sdk.BetaMessageNewParams {
	msgs := make([]sdk.BetaMessageParam, 0, len(req.Messages))
	for _, m := range req.Messages {
		role := sdk.BetaMessageParamRoleUser
		if m.Role == ai.RoleAssistant {
			role = sdk.BetaMessageParamRoleAssistant
		}
		msgs = append(msgs, sdk.BetaMessageParam{Role: role, Content: []sdk.BetaContentBlockParamUnion{sdk.NewBetaTextBlock(m.Content)}})
	}

	params := sdk.BetaMessageNewParams{
		Model:     p.model,
		MaxTokens: 64000,
		Messages:  msgs,
	}
	if req.System != "" {
		params.System = []sdk.BetaTextBlockParam{{Text: req.System}}
	}
	if p.fallbacks {
		params.Betas = []sdk.AnthropicBeta{sdk.AnthropicBetaServerSideFallback2026_07_01}
		params.Fallbacks = sdk.BetaFallbacksParamUnion{OfDefault: constant.ValueOf[constant.Default]()}
	}
	return params
}

func (p *Provider) run(ctx context.Context, params sdk.BetaMessageNewParams, onDelta func(string) error) (string, error) {
	stream := p.client.Beta.Messages.NewStreaming(ctx, params)
	defer stream.Close()

	var (
		out        strings.Builder
		stopReason sdk.BetaStopReason
	)
	for stream.Next() {
		switch ev := stream.Current().AsAny().(type) {
		case sdk.BetaRawContentBlockDeltaEvent:
			if delta, ok := ev.Delta.AsAny().(sdk.BetaTextDelta); ok {
				out.WriteString(delta.Text)
				if err := onDelta(delta.Text); err != nil {
					return "", err
				}
			}
		case sdk.BetaRawMessageDeltaEvent:
			stopReason = ev.Delta.StopReason
		}
	}
	if err := stream.Err(); err != nil {
		return "", fmt.Errorf("anthropic: %w", err)
	}
	if stopReason == sdk.BetaStopReasonRefusal {
		return "", ai.ErrRefused
	}
	return out.String(), nil
}
