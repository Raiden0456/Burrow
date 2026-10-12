package providers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/raiden0456/gonteractive/internal/ai"
	"github.com/raiden0456/gonteractive/internal/ai/acp"
	"github.com/raiden0456/gonteractive/internal/ai/anthropic"
	"github.com/raiden0456/gonteractive/internal/ai/openaicompat"
	"github.com/raiden0456/gonteractive/internal/httpx"
	"github.com/raiden0456/gonteractive/internal/settings"
	"github.com/raiden0456/gonteractive/internal/store/db"
)

type Kind string

const (
	KindAnthropic Kind = "anthropic"
	KindOpenAI    Kind = "openai"
	KindACP       Kind = "acp"
)

type Task string

const (
	TaskChat     Task = "chat"
	TaskGenerate Task = "generate"
	TaskGrade    Task = "grade"
)

const routingKey = "routing"

var ErrNotConfigured = &httpx.Error{
	Status:  http.StatusPreconditionFailed,
	Code:    "provider_not_configured",
	Message: "No AI provider configured. Add one in Settings.",
}

type Config struct {
	Model   string            `json:"model"`
	APIKey  string            `json:"apiKey"`
	BaseURL string            `json:"baseUrl"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

type Routing struct {
	Chat     int64 `json:"chat"`
	Generate int64 `json:"generate"`
	Grade    int64 `json:"grade"`
}

func (r Routing) For(task Task) int64 {
	switch task {
	case TaskChat:
		return r.Chat
	case TaskGenerate:
		return r.Generate
	default:
		return r.Grade
	}
}

type instance struct {
	kind     string
	config   string
	provider ai.Provider
}

type Service struct {
	q        *db.Queries
	settings *settings.Store

	mu   sync.Mutex
	live map[int64]instance
}

func New(q *db.Queries, s *settings.Store) *Service {
	return &Service{q: q, settings: s, live: map[int64]instance{}}
}

func (s *Service) For(ctx context.Context, task Task) (ai.Provider, error) {
	routing, err := s.Routing(ctx)
	if err != nil {
		return nil, err
	}

	var row db.Provider
	if id := routing.For(task); id != 0 {
		row, err = s.q.GetProvider(ctx, id)
	} else {
		row, err = s.first(ctx)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotConfigured
	}
	if err != nil {
		return nil, err
	}
	return s.instance(row)
}

func (s *Service) Routing(ctx context.Context) (Routing, error) {
	var r Routing
	_, err := s.settings.Get(ctx, routingKey, &r)
	return r, err
}

func (s *Service) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, inst := range s.live {
		inst.provider.Close()
		delete(s.live, id)
	}
}

func (s *Service) first(ctx context.Context) (db.Provider, error) {
	rows, err := s.q.ListProviders(ctx)
	if err != nil {
		return db.Provider{}, err
	}
	if len(rows) == 0 {
		return db.Provider{}, sql.ErrNoRows
	}
	return rows[0], nil
}

func (s *Service) instance(row db.Provider) (ai.Provider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if inst, ok := s.live[row.ID]; ok {
		if inst.kind == row.Kind && inst.config == row.Config {
			return inst.provider, nil
		}
		inst.provider.Close()
		delete(s.live, row.ID)
	}

	var cfg Config
	if err := json.Unmarshal([]byte(row.Config), &cfg); err != nil {
		return nil, fmt.Errorf("provider %q: decode config: %w", row.Name, err)
	}
	p, err := build(Kind(row.Kind), cfg)
	if err != nil {
		return nil, fmt.Errorf("provider %q: %w", row.Name, err)
	}
	s.live[row.ID] = instance{kind: row.Kind, config: row.Config, provider: p}
	return p, nil
}

func (s *Service) evict(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if inst, ok := s.live[id]; ok {
		inst.provider.Close()
		delete(s.live, id)
	}
}

func build(kind Kind, cfg Config) (ai.Provider, error) {
	switch kind {
	case KindAnthropic:
		if cfg.APIKey == "" {
			return nil, errors.New("API key is required")
		}
		return anthropic.New(cfg.APIKey, cfg.BaseURL, cfg.Model), nil
	case KindOpenAI:
		return openaicompat.New(cfg.APIKey, cfg.BaseURL, cfg.Model)
	case KindACP:
		return acp.New(cfg.Command, cfg.Args, cfg.Env)
	default:
		return nil, fmt.Errorf("unknown provider kind %q", kind)
	}
}
