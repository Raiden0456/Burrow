package lesson

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/raiden0456/gonteractive/internal/ai"
	"github.com/raiden0456/gonteractive/internal/curriculum"
	"github.com/raiden0456/gonteractive/internal/httpx"
	"github.com/raiden0456/gonteractive/internal/providers"
	"github.com/raiden0456/gonteractive/internal/settings"
	"github.com/raiden0456/gonteractive/internal/store/db"
	"github.com/raiden0456/gonteractive/internal/tutor"
)

const role = "You write lessons for a self-paced Go course."

const instructions = `Write a lesson that teaches this topic thoroughly but without filler.

Structure:
1. A short overview of what the topic is and why it matters.
2. Each learning objective as its own section with clear explanations and small, runnable code examples (show expected output in comments where useful).
3. "Gotchas": the mistakes people actually make with this topic, each with a minimal example.
4. "Idiomatic Go": how experienced Go developers approach it.
5. A compact cheat sheet at the end.

Aim for roughly 1000-1800 words. Start directly with the first heading; do not add a preamble.`

type Service struct {
	q          *db.Queries
	curriculum *curriculum.Curriculum
	providers  *providers.Service
	settings   *settings.Store
}

func New(q *db.Queries, c *curriculum.Curriculum, p *providers.Service, s *settings.Store) *Service {
	return &Service{q: q, curriculum: c, providers: p, settings: s}
}

func (s *Service) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/topics/{slug}/lesson", httpx.Handle(s.get))
	mux.HandleFunc("POST /api/topics/{slug}/lesson", httpx.Handle(s.generate))
}

func (s *Service) get(w http.ResponseWriter, r *http.Request) error {
	topic, err := s.curriculum.Lookup(r.PathValue("slug"))
	if err != nil {
		return err
	}
	l, err := s.q.GetLesson(r.Context(), topic.Slug)
	if errors.Is(err, sql.ErrNoRows) {
		httpx.JSON(w, http.StatusOK, map[string]any{"content": nil})
		return nil
	}
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"content": l.Content, "createdAt": l.CreatedAt})
	return nil
}

func (s *Service) generate(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	topic, err := s.curriculum.Lookup(r.PathValue("slug"))
	if err != nil {
		return err
	}
	prefs, err := s.settings.Preferences(ctx)
	if err != nil {
		return err
	}
	provider, err := s.providers.For(ctx, providers.TaskGenerate)
	if err != nil {
		return err
	}

	sse := httpx.NewSSE(w)
	content, err := provider.Stream(ctx, ai.Request{
		System:   tutor.System(prefs, role),
		Messages: []ai.Message{{Role: ai.RoleUser, Content: topic.Brief() + "\n" + instructions}},
	}, sse.Delta)
	if err != nil {
		sse.Fail(err)
		return nil
	}
	if err := s.q.UpsertLesson(ctx, db.UpsertLessonParams{TopicSlug: topic.Slug, Content: content}); err != nil {
		sse.Fail(err)
		return nil
	}
	sse.Send("done", map[string]any{"content": content})
	return nil
}
