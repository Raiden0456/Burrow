package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/raiden0456/gonteractive/internal/ai"
	"github.com/raiden0456/gonteractive/internal/httpx"
	"github.com/raiden0456/gonteractive/internal/store/db"
)

type providerView struct {
	ID        int64             `json:"id"`
	Name      string            `json:"name"`
	Kind      Kind              `json:"kind"`
	Model     string            `json:"model"`
	BaseURL   string            `json:"baseUrl"`
	Command   string            `json:"command"`
	Args      []string          `json:"args"`
	Env       map[string]string `json:"env"`
	HasAPIKey bool              `json:"hasApiKey"`
}

type providerInput struct {
	Name   string `json:"name"`
	Kind   Kind   `json:"kind"`
	Config Config `json:"config"`
}

var pingSchema = ai.MustSchema("ping", []byte(`{
	"type": "object",
	"properties": {"reply": {"type": "string"}},
	"required": ["reply"],
	"additionalProperties": false
}`))

func (s *Service) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/providers", httpx.Handle(s.list))
	mux.HandleFunc("POST /api/providers", httpx.Handle(s.create))
	mux.HandleFunc("PUT /api/providers/{id}", httpx.Handle(s.update))
	mux.HandleFunc("DELETE /api/providers/{id}", httpx.Handle(s.delete))
	mux.HandleFunc("POST /api/providers/{id}/test", httpx.Handle(s.test))
	mux.HandleFunc("GET /api/routing", httpx.Handle(s.getRouting))
	mux.HandleFunc("PUT /api/routing", httpx.Handle(s.putRouting))
}

func (s *Service) list(w http.ResponseWriter, r *http.Request) error {
	rows, err := s.q.ListProviders(r.Context())
	if err != nil {
		return err
	}
	views := make([]providerView, 0, len(rows))
	for _, row := range rows {
		v, err := toView(row)
		if err != nil {
			return err
		}
		views = append(views, v)
	}
	httpx.JSON(w, http.StatusOK, views)
	return nil
}

func (s *Service) create(w http.ResponseWriter, r *http.Request) error {
	in, err := decodeInput(r)
	if err != nil {
		return err
	}
	raw, _ := json.Marshal(in.Config)
	row, err := s.q.CreateProvider(r.Context(), db.CreateProviderParams{Name: in.Name, Kind: string(in.Kind), Config: string(raw)})
	if err != nil {
		return conflictOr(err)
	}
	v, err := toView(row)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, v)
	return nil
}

func (s *Service) update(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	in, err := decodeInput(r)
	if err != nil {
		return err
	}
	existing, err := s.q.GetProvider(r.Context(), id)
	if err != nil {
		return err
	}
	if in.Config.APIKey == "" {
		var old Config
		json.Unmarshal([]byte(existing.Config), &old)
		in.Config.APIKey = old.APIKey
	}
	raw, _ := json.Marshal(in.Config)
	row, err := s.q.UpdateProvider(r.Context(), db.UpdateProviderParams{ID: id, Name: in.Name, Kind: string(in.Kind), Config: string(raw)})
	if err != nil {
		return conflictOr(err)
	}
	s.evict(id)
	v, err := toView(row)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, v)
	return nil
}

func (s *Service) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	if err := s.q.DeleteProvider(r.Context(), id); err != nil {
		return err
	}
	s.evict(id)

	routing, err := s.Routing(r.Context())
	if err != nil {
		return err
	}
	for _, slot := range []*int64{&routing.Chat, &routing.Generate, &routing.Grade} {
		if *slot == id {
			*slot = 0
		}
	}
	if err := s.settings.Put(r.Context(), routingKey, routing); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Service) test(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	row, err := s.q.GetProvider(r.Context(), id)
	if err != nil {
		return err
	}
	p, err := s.instance(row)
	if err != nil {
		return httpx.Errorf(http.StatusBadRequest, "%v", err)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
	defer cancel()
	start := time.Now()
	var out struct {
		Reply string `json:"reply"`
	}
	err = ai.GenerateJSON(ctx, p, ai.Request{
		Messages: []ai.Message{{Role: ai.RoleUser, Content: `Connection test. Set "reply" to a one-sentence greeting for a Go learner.`}},
	}, pingSchema, &out)
	if err != nil {
		httpx.JSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return nil
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true, "reply": out.Reply, "latencyMs": time.Since(start).Milliseconds()})
	return nil
}

func (s *Service) getRouting(w http.ResponseWriter, r *http.Request) error {
	routing, err := s.Routing(r.Context())
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, routing)
	return nil
}

func (s *Service) putRouting(w http.ResponseWriter, r *http.Request) error {
	var routing Routing
	if err := httpx.Decode(r, &routing); err != nil {
		return err
	}
	for _, id := range []int64{routing.Chat, routing.Generate, routing.Grade} {
		if id == 0 {
			continue
		}
		if _, err := s.q.GetProvider(r.Context(), id); err != nil {
			return httpx.Errorf(http.StatusBadRequest, "unknown provider %d", id)
		}
	}
	if err := s.settings.Put(r.Context(), routingKey, routing); err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, routing)
	return nil
}

func decodeInput(r *http.Request) (providerInput, error) {
	var in providerInput
	if err := httpx.Decode(r, &in); err != nil {
		return in, err
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return in, httpx.Errorf(http.StatusBadRequest, "name is required")
	}
	switch in.Kind {
	case KindAnthropic, KindOpenAI:
		in.Config.Command, in.Config.Args, in.Config.Env = "", nil, nil
	case KindACP:
		in.Config.APIKey, in.Config.BaseURL = "", ""
		if strings.TrimSpace(in.Config.Command) == "" {
			return in, httpx.Errorf(http.StatusBadRequest, "command is required for ACP providers")
		}
	default:
		return in, httpx.Errorf(http.StatusBadRequest, "kind must be anthropic, openai or acp")
	}
	if in.Kind == KindOpenAI && in.Config.Model == "" {
		return in, httpx.Errorf(http.StatusBadRequest, "model is required for OpenAI-compatible providers")
	}
	return in, nil
}

func toView(row db.Provider) (providerView, error) {
	var cfg Config
	if err := json.Unmarshal([]byte(row.Config), &cfg); err != nil {
		return providerView{}, err
	}
	return providerView{
		ID:        row.ID,
		Name:      row.Name,
		Kind:      Kind(row.Kind),
		Model:     cfg.Model,
		BaseURL:   cfg.BaseURL,
		Command:   cfg.Command,
		Args:      cfg.Args,
		Env:       cfg.Env,
		HasAPIKey: cfg.APIKey != "",
	}, nil
}

func conflictOr(err error) error {
	if strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return httpx.Errorf(http.StatusConflict, "a provider with this name already exists")
	}
	return err
}
