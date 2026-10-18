package daily

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/raiden0456/gonteractive/internal/httpx"
	"github.com/raiden0456/gonteractive/internal/jobs"
	"github.com/raiden0456/gonteractive/internal/store/db"
)

func (s *Scheduler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/daily", httpx.Handle(s.getToday))
	mux.HandleFunc("POST /api/daily", httpx.Handle(s.createToday))
	mux.HandleFunc("GET /api/push/public-key", httpx.Handle(s.publicKey))
	mux.HandleFunc("POST /api/push/subscriptions", httpx.Handle(s.subscribe))
	mux.HandleFunc("DELETE /api/push/subscriptions", httpx.Handle(s.unsubscribe))
	mux.HandleFunc("POST /api/push/test", httpx.Handle(s.testPush))
}

func (s *Scheduler) getToday(w http.ResponseWriter, r *http.Request) error {
	now := time.Now()
	q, err := s.Today(r.Context(), now)
	if err != nil {
		return err
	}
	prefs, err := s.settings.Preferences(r.Context())
	if err != nil {
		return err
	}
	var quizID *int64
	if q != nil {
		quizID = &q.ID
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"date":        now.Format(dateLayout),
		"quizId":      quizID,
		"enabled":     prefs.DailyEnabled,
		"scheduledAt": prefs.DailyAt(now).Unix(),
	})
	return nil
}

func (s *Scheduler) createToday(w http.ResponseWriter, r *http.Request) error {
	prefs, err := s.settings.Preferences(r.Context())
	if err != nil {
		return err
	}
	job, err := s.jobs.Start(r.Context(), jobs.KindDaily, "", func(ctx context.Context, _ func(string)) (int64, error) {
		q, _, err := s.Ensure(ctx, time.Now(), prefs.DailySize)
		return q.ID, err
	})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusAccepted, job)
	return nil
}

func (s *Scheduler) publicKey(w http.ResponseWriter, r *http.Request) error {
	count, err := s.q.CountPushSubscriptions(r.Context())
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"publicKey": s.pusher.PublicKey(), "subscriptions": count})
	return nil
}

func (s *Scheduler) subscribe(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Endpoint string `json:"endpoint"`
		Keys     struct {
			P256dh string `json:"p256dh"`
			Auth   string `json:"auth"`
		} `json:"keys"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	if !strings.HasPrefix(in.Endpoint, "https://") || in.Keys.P256dh == "" || in.Keys.Auth == "" {
		return httpx.Errorf(http.StatusBadRequest, "invalid push subscription")
	}
	err := s.q.UpsertPushSubscription(r.Context(), db.UpsertPushSubscriptionParams{Endpoint: in.Endpoint, P256dh: in.Keys.P256dh, Auth: in.Keys.Auth})
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Scheduler) unsubscribe(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Endpoint string `json:"endpoint"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	if err := s.q.DeletePushSubscription(r.Context(), in.Endpoint); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Scheduler) testPush(w http.ResponseWriter, r *http.Request) error {
	sent, err := s.pusher.Notify(r.Context(), Notification{
		Title: "Burrow",
		Body:  s.messages(r.Context()).testBody,
		URL:   "/",
	})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, map[string]int{"sent": sent})
	return nil
}
