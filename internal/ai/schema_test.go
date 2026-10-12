package ai

import (
	"context"
	"errors"
	"testing"
)

type scriptedProvider struct {
	replies []string
	calls   []Request
}

func (s *scriptedProvider) Stream(context.Context, Request, func(string) error) (string, error) {
	return "", errors.New("not used")
}

func (s *scriptedProvider) Generate(_ context.Context, req Request, _ *Schema) (string, error) {
	s.calls = append(s.calls, req)
	reply := s.replies[0]
	s.replies = s.replies[1:]
	return reply, nil
}

func (s *scriptedProvider) Close() error { return nil }

var countSchema = MustSchema("count", []byte(`{"type":"object","properties":{"n":{"type":"integer"}},"required":["n"],"additionalProperties":false}`))

type count struct {
	N int `json:"n"`
}

func (c *count) Validate() error {
	if c.N < 1 {
		return errors.New("n must be positive")
	}
	return nil
}

func TestGenerateJSONAcceptsFencedJSON(t *testing.T) {
	p := &scriptedProvider{replies: []string{"Sure!\n```json\n{\"n\": 3}\n```"}}
	var out count
	if err := GenerateJSON(t.Context(), p, Request{Messages: []Message{{Role: RoleUser, Content: "go"}}}, countSchema, &out); err != nil {
		t.Fatal(err)
	}
	if out.N != 3 || len(p.calls) != 1 {
		t.Fatalf("got n=%d after %d calls", out.N, len(p.calls))
	}
}

func TestGenerateJSONRetriesWithValidatorError(t *testing.T) {
	p := &scriptedProvider{replies: []string{`{"n": 0}`, `{"n": 2}`}}
	var out count
	if err := GenerateJSON(t.Context(), p, Request{Messages: []Message{{Role: RoleUser, Content: "go"}}}, countSchema, &out); err != nil {
		t.Fatal(err)
	}
	if out.N != 2 {
		t.Fatalf("got n=%d, want 2", out.N)
	}
	retry := p.calls[1].Messages
	if len(retry) != 3 || retry[2].Role != RoleUser {
		t.Fatalf("retry should append the bad reply and a correction request, got %+v", retry)
	}
}

func TestGenerateJSONGivesUpAfterSecondFailure(t *testing.T) {
	p := &scriptedProvider{replies: []string{`{"x": 1}`, `not json`}}
	var out count
	if err := GenerateJSON(t.Context(), p, Request{}, countSchema, &out); err == nil {
		t.Fatal("expected an error")
	}
}
