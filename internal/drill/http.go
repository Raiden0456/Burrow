package drill

import (
	"net/http"

	"github.com/raiden0456/gonteractive/internal/httpx"
	"github.com/raiden0456/gonteractive/internal/store/db"
)

type summaryView struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
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
	ID          int64        `json:"id"`
	TopicSlug   string       `json:"topicSlug"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	StarterCode string       `json:"starterCode"`
	Difficulty  int64        `json:"difficulty"`
	Solved      bool         `json:"solved"`
	LastAttempt *attemptView `json:"lastAttempt"`
}

func (s *Service) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/topics/{slug}/drills", httpx.Handle(s.list))
	mux.HandleFunc("POST /api/topics/{slug}/drills", httpx.Handle(s.create))
	mux.HandleFunc("GET /api/drills/{id}", httpx.Handle(s.get))
	mux.HandleFunc("GET /api/drills/{id}/solution", httpx.Handle(s.solution))
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
		out = append(out, summaryView{ID: d.ID, Title: d.Title, Difficulty: d.Difficulty, Solved: d.SolvedAt.Valid, CreatedAt: d.CreatedAt})
	}
	httpx.JSON(w, http.StatusOK, out)
	return nil
}

func (s *Service) create(w http.ResponseWriter, r *http.Request) error {
	d, err := s.Generate(r.Context(), r.PathValue("slug"))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, map[string]int64{"id": d.ID})
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
	httpx.JSON(w, http.StatusOK, toView(d, attempt))
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
	httpx.JSON(w, http.StatusOK, map[string]string{"solution": d.Solution, "testCode": d.TestCode})
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
		Difficulty:  d.Difficulty,
		Solved:      d.SolvedAt.Valid,
	}
	if a != nil {
		v.LastAttempt = &attemptView{Code: a.Code, Passed: a.Passed == 1, Output: a.Output, CreatedAt: a.CreatedAt}
	}
	return v
}
