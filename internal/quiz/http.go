package quiz

import (
	"net/http"
	"strconv"

	"github.com/raiden0456/gonteractive/internal/httpx"
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
	mux.HandleFunc("POST /api/quizzes/review", httpx.Handle(s.createReview))
	mux.HandleFunc("GET /api/quizzes", httpx.Handle(s.list))
	mux.HandleFunc("GET /api/quizzes/{id}", httpx.Handle(s.get))
	mux.HandleFunc("POST /api/quizzes/{id}/answers", httpx.Handle(s.answer))
}

func (s *Service) createTopic(w http.ResponseWriter, r *http.Request) error {
	var in sizeInput
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	quiz, err := s.CreateTopicQuiz(r.Context(), r.PathValue("slug"), in.clamp(5))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, map[string]int64{"id": quiz.ID})
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
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	out, err := s.Recent(r.Context(), limit)
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
