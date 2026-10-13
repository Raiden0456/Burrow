package jobs

import (
	"net/http"
	"time"

	"github.com/raiden0456/gonteractive/internal/httpx"
)

func (m *Manager) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/jobs", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		jobs, err := m.Recent(r.Context(), time.Now())
		if err != nil {
			return err
		}
		httpx.JSON(w, http.StatusOK, jobs)
		return nil
	}))

	mux.HandleFunc("GET /api/jobs/{id}", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		id, err := httpx.PathID(r, "id")
		if err != nil {
			return err
		}
		job, err := m.Get(r.Context(), id)
		if err != nil {
			return err
		}
		httpx.JSON(w, http.StatusOK, job)
		return nil
	}))
}
