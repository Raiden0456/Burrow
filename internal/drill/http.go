package drill

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/raiden0456/gonteractive/internal/httpx"
	"github.com/raiden0456/gonteractive/internal/jobs"
	"github.com/raiden0456/gonteractive/internal/providers"
	"github.com/raiden0456/gonteractive/internal/store/db"
)

type summaryView struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Mode       Mode   `json:"mode"`
	Difficulty int64  `json:"difficulty"`
	Solved     bool   `json:"solved"`
	CreatedAt  int64  `json:"createdAt"`
}

type attemptView struct {
	Code      string `json:"code"`
	Passed    bool   `json:"passed"`
	Output    string `json:"output"`
	CreatedAt int64  `json:"createdAt"`
}

type drillView struct {
	ID            int64        `json:"id"`
	TopicSlug     string       `json:"topicSlug"`
	Title         string       `json:"title"`
	Description   string       `json:"description"`
	StarterCode   string       `json:"starterCode"`
	Mode          Mode         `json:"mode"`
	TestCode      string       `json:"testCode,omitempty"`
	InitialOutput string       `json:"initialOutput,omitempty"`
	Difficulty    int64        `json:"difficulty"`
	Solved        bool         `json:"solved"`
	Draft         string       `json:"draft"`
	LastAttempt   *attemptView `json:"lastAttempt"`
	Hints         []Hint       `json:"hints"`
}

func (s *Service) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/topics/{slug}/drills", httpx.Handle(s.list))
	mux.HandleFunc("POST /api/topics/{slug}/drills", httpx.Handle(s.create))
	mux.HandleFunc("GET /api/drills/{id}", httpx.Handle(s.get))
	mux.HandleFunc("GET /api/drills/{id}/solution", httpx.Handle(s.solution))
	mux.HandleFunc("PUT /api/drills/{id}/draft", httpx.Handle(s.saveDraft))
	mux.HandleFunc("POST /api/drills/{id}/help", httpx.Handle(s.help))
	mux.HandleFunc("POST /api/drills/{id}/run", httpx.Handle(s.run))
}

func (s *Service) list(w http.ResponseWriter, r *http.Request) error {
	topic, err := s.curriculum.Lookup(r.PathValue("slug"))
	if err != nil {
		return err
	}
	rows, err := s.q.ListDrillsByTopic(r.Context(), topic.Slug)
	if err != nil {
		return err
	}
	out := make([]summaryView, 0, len(rows))
	for _, d := range rows {
		out = append(out, summaryView{ID: d.ID, Title: d.Title, Mode: Mode(d.Mode), Difficulty: d.Difficulty, Solved: d.SolvedAt.Valid, CreatedAt: d.CreatedAt})
	}
	httpx.JSON(w, http.StatusOK, out)
	return nil
}

func (s *Service) create(w http.ResponseWriter, r *http.Request) error {
	mode := Mode(r.URL.Query().Get("mode"))
	if mode == "" {
		mode = ModeImplement
	}
	if mode != ModeImplement && mode != ModeBugHunt {
		return httpx.Coded(http.StatusBadRequest, "invalid_drill_mode", "invalid drill mode")
	}
	topic, err := s.curriculum.Lookup(r.PathValue("slug"))
	if err != nil {
		return err
	}
	if s.runner == nil {
		return s.runnerErr
	}
	if _, err := s.providers.For(r.Context(), providers.TaskGenerate); err != nil {
		return err
	}
	job, err := s.jobs.Start(r.Context(), jobs.KindDrill, topic.Slug, func(ctx context.Context, _ func(string)) (int64, error) {
		d, err := s.Generate(ctx, topic.Slug, mode)
		return d.ID, err
	})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusAccepted, job)
	return nil
}

func (s *Service) get(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	d, err := s.q.GetDrill(r.Context(), id)
	if err != nil {
		return err
	}
	attempt, err := s.latestAttempt(r.Context(), id)
	if err != nil {
		return err
	}
	hints, err := s.q.ListDrillHints(r.Context(), sql.NullInt64{Int64: id, Valid: true})
	if err != nil {
		return err
	}
	view := toView(d, attempt)
	for _, h := range hints {
		view.Hints = append(view.Hints, Hint{ID: h.ID, Content: h.Content, CreatedAt: h.CreatedAt})
	}
	httpx.JSON(w, http.StatusOK, view)
	return nil
}

func (s *Service) saveDraft(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in struct {
		Code string `json:"code"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	n, err := s.q.SaveDrillDraft(r.Context(), db.SaveDrillDraftParams{ID: id, Draft: in.Code})
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Service) help(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in struct {
		Code   string `json:"code"`
		Output string `json:"output"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	if _, err := s.q.GetDrill(r.Context(), id); err != nil {
		return err
	}
	if _, err := s.providers.For(r.Context(), providers.TaskChat); err != nil {
		return err
	}

	sse := httpx.NewSSE(w)
	hint, err := s.StreamHelp(r.Context(), id, in.Code, in.Output, sse.Delta)
	if err != nil {
		sse.Fail(err)
		return nil
	}
	sse.Send("done", hint)
	return nil
}

func (s *Service) solution(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	d, err := s.q.GetDrill(r.Context(), id)
	if err != nil {
		return err
	}
	if Mode(d.Mode) == ModeBugHunt && !d.SolvedAt.Valid {
		return httpx.Coded(http.StatusForbidden, "bug_hunt_unsolved", "pass the tests to unlock the Bug Hunt explanation")
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"solution": d.Solution, "testCode": d.TestCode, "explanation": d.Explanation})
	return nil
}

func (s *Service) run(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in struct {
		Code string `json:"code"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	res, err := s.Run(r.Context(), id, in.Code)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, res)
	return nil
}

func toView(d db.Drill, a *db.DrillAttempt) drillView {
	v := drillView{
		ID:          d.ID,
		TopicSlug:   d.TopicSlug,
		Title:       d.Title,
		Description: d.Description,
		StarterCode: d.StarterCode,
		Mode:        Mode(d.Mode),
		Difficulty:  d.Difficulty,
		Solved:      d.SolvedAt.Valid,
		Draft:       d.Draft,
		Hints:       []Hint{},
	}
	if Mode(d.Mode) == ModeBugHunt {
		v.TestCode = d.TestCode
		v.InitialOutput = d.InitialOutput
	}
	if a != nil {
		v.LastAttempt = &attemptView{Code: a.Code, Passed: a.Passed == 1, Output: a.Output, CreatedAt: a.CreatedAt}
	}
	return v
}
