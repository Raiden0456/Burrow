package quiz

import (
	"context"
	"net/http"
	"strconv"

	"github.com/raiden0456/gonteractive/internal/httpx"
	"github.com/raiden0456/gonteractive/internal/jobs"
	"github.com/raiden0456/gonteractive/internal/providers"
)

type sizeInput struct {
	Size int `json:"size"`
}

func (in sizeInput) clamp(fallback int) int {
	if in.Size <= 0 {
		return fallback
	}
	return min(in.Size, 20)
}

func (s *Service) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/topics/{slug}/quizzes", httpx.Handle(s.createTopic))
	mux.HandleFunc("POST /api/topics/{slug}/checkpoint", httpx.Handle(s.createCheckpoint))
	mux.HandleFunc("POST /api/quizzes/review", httpx.Handle(s.createReview))
	mux.HandleFunc("GET /api/quizzes", httpx.Handle(s.list))
	mux.HandleFunc("GET /api/quizzes/{id}", httpx.Handle(s.get))
	mux.HandleFunc("POST /api/quizzes/{id}/answers", httpx.Handle(s.answer))
	mux.HandleFunc("PUT /api/quizzes/{id}/drafts", httpx.Handle(s.saveDraft))
	mux.HandleFunc("POST /api/quizzes/{id}/questions/{qid}/hints", httpx.Handle(s.hint))
}

func (s *Service) createTopic(w http.ResponseWriter, r *http.Request) error {
	var in sizeInput
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	topic, err := s.curriculum.Lookup(r.PathValue("slug"))
	if err != nil {
		return err
	}
	if _, err := s.providers.For(r.Context(), providers.TaskGenerate); err != nil {
		return err
	}
	size := in.clamp(5)
	job, err := s.jobs.Start(r.Context(), jobs.KindQuiz, topic.Slug, func(ctx context.Context, _ func(string)) (int64, error) {
		quiz, err := s.CreateTopicQuiz(ctx, topic.Slug, size)
		return quiz.ID, err
	})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusAccepted, job)
	return nil
}

func (s *Service) createCheckpoint(w http.ResponseWriter, r *http.Request) error {
	topic, err := s.curriculum.Lookup(r.PathValue("slug"))
	if err != nil {
		return err
	}
	if _, err := s.providers.For(r.Context(), providers.TaskGenerate); err != nil {
		return err
	}
	job, err := s.jobs.Start(r.Context(), jobs.KindCheckpoint, topic.Slug, func(ctx context.Context, _ func(string)) (int64, error) {
		quiz, err := s.CreateCheckpointQuiz(ctx, topic.Slug)
		return quiz.ID, err
	})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusAccepted, job)
	return nil
}

func (s *Service) createReview(w http.ResponseWriter, r *http.Request) error {
	var in sizeInput
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	quiz, err := s.CreateReviewQuiz(r.Context(), in.clamp(10))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, map[string]int64{"id": quiz.ID})
	return nil
}

func (s *Service) list(w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()
	f := Filter{TopicSlug: query.Get("topic")}
	f.Limit, _ = strconv.ParseInt(query.Get("limit"), 10, 64)
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 20
	}
	switch query.Get("status") {
	case "open":
		open := true
		f.Open = &open
	case "done":
		open := false
		f.Open = &open
	}
	out, err := s.List(r.Context(), f)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, out)
	return nil
}

func (s *Service) get(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	view, err := s.View(r.Context(), id)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, view)
	return nil
}

func (s *Service) answer(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in struct {
		QuestionID int64    `json:"questionId"`
		Response   Response `json:"response"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	res, completed, err := s.Answer(r.Context(), id, in.QuestionID, in.Response)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"result": res, "quizCompleted": completed})
	return nil
}

func (s *Service) saveDraft(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in struct {
		QuestionID int64    `json:"questionId"`
		Response   Response `json:"response"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	if err := s.SaveDraft(r.Context(), id, in.QuestionID, in.Response); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Service) hint(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	qid, err := httpx.PathID(r, "qid")
	if err != nil {
		return err
	}
	var in struct {
		Kind string `json:"kind"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	if _, err := s.loadHintRequest(r.Context(), id, qid); err != nil {
		return err
	}
	if _, err := s.providers.For(r.Context(), providers.TaskChat); err != nil {
		return err
	}

	sse := httpx.NewSSE(w)
	hint, err := s.StreamHint(r.Context(), id, qid, in.Kind, sse.Delta)
	if err != nil {
		sse.Fail(err)
		return nil
	}
	sse.Send("done", hint)
	return nil
}
