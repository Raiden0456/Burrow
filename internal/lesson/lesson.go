package lesson

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/raiden0456/gonteractive/internal/ai"
	"github.com/raiden0456/gonteractive/internal/curriculum"
	"github.com/raiden0456/gonteractive/internal/httpx"
	"github.com/raiden0456/gonteractive/internal/jobs"
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
	jobs       *jobs.Manager
}

func New(q *db.Queries, c *curriculum.Curriculum, p *providers.Service, s *settings.Store, j *jobs.Manager) *Service {
	return &Service{q: q, curriculum: c, providers: p, settings: s, jobs: j}
}

type lessonView struct {
	Content   *string   `json:"content"`
	CreatedAt *int64    `json:"createdAt"`
	Job       *jobs.Job `json:"job"`
}

type lessonSummary struct {
	TopicSlug string `json:"topicSlug"`
	CreatedAt int64  `json:"createdAt"`
}

func (s *Service) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/lessons", httpx.Handle(s.list))
	mux.HandleFunc("GET /api/topics/{slug}/lesson", httpx.Handle(s.get))
	mux.HandleFunc("POST /api/topics/{slug}/lesson", httpx.Handle(s.generate))
}

func (s *Service) list(w http.ResponseWriter, r *http.Request) error {
	rows, err := s.q.ListLessons(r.Context())
	if err != nil {
		return err
	}
	out := make([]lessonSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, lessonSummary{TopicSlug: row.TopicSlug, CreatedAt: row.CreatedAt})
	}
	httpx.JSON(w, http.StatusOK, out)
	return nil
}

func (s *Service) get(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	topic, err := s.curriculum.Lookup(r.PathValue("slug"))
	if err != nil {
		return err
	}
	var view lessonView
	l, err := s.q.GetLesson(ctx, topic.Slug)
	switch {
	case err == nil:
		view.Content, view.CreatedAt = &l.Content, &l.CreatedAt
	case !errors.Is(err, sql.ErrNoRows):
		return err
	}
	if view.Job, err = s.jobs.Running(ctx, jobs.KindLesson, topic.Slug); err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, view)
	return nil
}

func (s *Service) generate(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	topic, err := s.curriculum.Lookup(r.PathValue("slug"))
	if err != nil {
		return err
	}
	if _, err := s.providers.For(ctx, providers.TaskGenerate); err != nil {
		return err
	}
	job, err := s.jobs.Start(ctx, jobs.KindLesson, topic.Slug, func(ctx context.Context, progress func(string)) (int64, error) {
		return 0, s.write(ctx, topic, progress)
	})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusAccepted, job)
	return nil
}

func (s *Service) write(ctx context.Context, topic curriculum.Topic, progress func(string)) error {
	prefs, err := s.settings.Preferences(ctx)
	if err != nil {
		return err
	}
	provider, err := s.providers.For(ctx, providers.TaskGenerate)
	if err != nil {
		return err
	}
	content, err := provider.Stream(ctx, ai.Request{
		System:   tutor.System(prefs, role),
		Messages: []ai.Message{{Role: ai.RoleUser, Content: topic.Brief() + "\n" + instructions}},
	}, func(delta string) error {
		progress(delta)
		return nil
	})
	if err != nil {
		return err
	}
	return s.q.UpsertLesson(ctx, db.UpsertLessonParams{TopicSlug: topic.Slug, Content: content})
}
