package daily

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/raiden0456/gonteractive/internal/providers"
	"github.com/raiden0456/gonteractive/internal/quiz"
	"github.com/raiden0456/gonteractive/internal/settings"
	"github.com/raiden0456/gonteractive/internal/store/db"
)

const (
	dateLayout    = "2006-01-02"
	retryAfter    = 30 * time.Minute
	buildDeadline = 10 * time.Minute
)

type Scheduler struct {
	q        *db.Queries
	settings *settings.Store
	quizzes  *quiz.Service
	pusher   *Pusher

	mu          sync.Mutex
	lastFailure time.Time
}

func NewScheduler(q *db.Queries, s *settings.Store, quizzes *quiz.Service, pusher *Pusher) *Scheduler {
	return &Scheduler{q: q, settings: s, quizzes: quizzes, pusher: pusher}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		s.tick(ctx, time.Now())
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Scheduler) tick(ctx context.Context, now time.Time) {
	prefs, err := s.settings.Preferences(ctx)
	if err != nil {
		slog.Error("daily: load preferences", "err", err)
		return
	}
	if !prefs.DailyEnabled || now.Before(prefs.DailyAt(now)) {
		return
	}
	if !s.lastFailure.IsZero() && now.Sub(s.lastFailure) < retryAfter {
		return
	}

	buildCtx, cancel := context.WithTimeout(ctx, buildDeadline)
	defer cancel()
	q, created, err := s.Ensure(buildCtx, now, prefs.DailySize)
	if err != nil {
		s.lastFailure = now
		if errors.Is(err, providers.ErrNotConfigured) {
			slog.Warn("daily: skipped, no AI provider configured")
			return
		}
		slog.Error("daily: build quiz", "err", err)
		return
	}
	s.lastFailure = time.Time{}
	if !created {
		return
	}

	slog.Info("daily: quiz ready", "quiz", q.ID)
	msg := s.messages(ctx)
	if _, err := s.pusher.Notify(ctx, Notification{
		Title: msg.dailyTitle,
		Body:  msg.dailyBody,
		URL:   fmt.Sprintf("/quiz/%d", q.ID),
	}); err != nil {
		slog.Error("daily: notify", "err", err)
	}
}

func (s *Scheduler) Ensure(ctx context.Context, now time.Time, size int) (db.Quiz, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	date := now.Format(dateLayout)
	existing, err := s.q.GetDailyQuiz(ctx, sql.NullString{String: date, Valid: true})
	if err == nil {
		return existing, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return db.Quiz{}, false, err
	}
	q, err := s.quizzes.CreateDailyQuiz(ctx, date, size)
	return q, err == nil, err
}

func (s *Scheduler) Today(ctx context.Context, now time.Time) (*db.Quiz, error) {
	q, err := s.q.GetDailyQuiz(ctx, sql.NullString{String: now.Format(dateLayout), Valid: true})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &q, nil
}
