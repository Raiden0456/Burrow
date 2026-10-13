package jobs

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/raiden0456/gonteractive/internal/store/db"
)

type Kind string

const (
	KindLesson Kind = "lesson"
	KindQuiz   Kind = "quiz"
	KindDrill  Kind = "drill"
	KindDaily  Kind = "daily"
)

const (
	statusRunning = "running"
	statusDone    = "done"
	statusFailed  = "failed"

	jobTimeout    = 20 * time.Minute
	recentWindow  = 15 * time.Minute
	maxConcurrent = 3
)

type Job struct {
	ID         int64  `json:"id"`
	Kind       Kind   `json:"kind"`
	TopicSlug  string `json:"topicSlug"`
	Status     string `json:"status"`
	ResultID   *int64 `json:"resultId"`
	Error      string `json:"error"`
	CreatedAt  int64  `json:"createdAt"`
	FinishedAt *int64 `json:"finishedAt"`
	Partial    string `json:"partial,omitempty"`
}

type Work func(ctx context.Context, progress func(string)) (resultID int64, err error)

type Manager struct {
	q     *db.Queries
	base  context.Context
	slots chan struct{}
	wg    sync.WaitGroup

	mu      sync.Mutex
	starts  sync.Mutex
	partial map[int64]*strings.Builder
}

// New returns a manager whose jobs run on base, independent of any HTTP request. Jobs that were
// still running when the previous process stopped are marked as failed.
func New(base context.Context, q *db.Queries) (*Manager, error) {
	if err := q.FailRunningJobs(base, "interrupted by a server restart"); err != nil {
		return nil, err
	}
	return &Manager{q: q, base: base, slots: make(chan struct{}, maxConcurrent), partial: map[int64]*strings.Builder{}}, nil
}

// Start runs work in the background unless a job of the same kind and topic is already running, in
// which case that job is returned instead.
func (m *Manager) Start(ctx context.Context, kind Kind, topic string, work Work) (Job, error) {
	m.starts.Lock()
	defer m.starts.Unlock()

	existing, err := m.q.FindRunningJob(ctx, db.FindRunningJobParams{Kind: string(kind), TopicSlug: topic})
	if err == nil {
		return m.view(existing), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Job{}, err
	}

	row, err := m.q.CreateJob(ctx, db.CreateJobParams{Kind: string(kind), TopicSlug: topic})
	if err != nil {
		return Job{}, err
	}
	m.mu.Lock()
	m.partial[row.ID] = &strings.Builder{}
	m.mu.Unlock()

	m.wg.Add(1)
	go m.run(row.ID, work)
	return m.view(row), nil
}

func (m *Manager) run(id int64, work Work) {
	defer m.wg.Done()
	defer func() {
		m.mu.Lock()
		delete(m.partial, id)
		m.mu.Unlock()
	}()

	ctx, cancel := context.WithTimeout(m.base, jobTimeout)
	defer cancel()

	select {
	case m.slots <- struct{}{}:
		defer func() { <-m.slots }()
	case <-ctx.Done():
		m.finish(id, 0, ctx.Err())
		return
	}

	resultID, err := work(ctx, func(chunk string) {
		m.mu.Lock()
		defer m.mu.Unlock()
		if b, ok := m.partial[id]; ok {
			b.WriteString(chunk)
		}
	})
	m.finish(id, resultID, err)
}

func (m *Manager) finish(id, resultID int64, err error) {
	params := db.FinishJobParams{ID: id, Status: statusDone}
	if err != nil {
		params.Status = statusFailed
		params.Error = err.Error()
		slog.Warn("job failed", "job", id, "err", err)
	} else if resultID != 0 {
		params.ResultID = sql.NullInt64{Int64: resultID, Valid: true}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.q.FinishJob(ctx, params); err != nil {
		slog.Error("finish job", "job", id, "err", err)
	}
}

func (m *Manager) Get(ctx context.Context, id int64) (Job, error) {
	row, err := m.q.GetJob(ctx, id)
	if err != nil {
		return Job{}, err
	}
	return m.view(row), nil
}

func (m *Manager) Running(ctx context.Context, kind Kind, topic string) (*Job, error) {
	row, err := m.q.FindRunningJob(ctx, db.FindRunningJobParams{Kind: string(kind), TopicSlug: topic})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	job := m.view(row)
	return &job, nil
}

func (m *Manager) Recent(ctx context.Context, now time.Time) ([]Job, error) {
	rows, err := m.q.ListRecentJobs(ctx, sql.NullInt64{Int64: now.Add(-recentWindow).Unix(), Valid: true})
	if err != nil {
		return nil, err
	}
	out := make([]Job, 0, len(rows))
	for _, r := range rows {
		job := m.view(r)
		job.Partial = ""
		out = append(out, job)
	}
	return out, nil
}

func (m *Manager) Wait() {
	m.wg.Wait()
}

func (m *Manager) view(r db.Job) Job {
	job := Job{
		ID:        r.ID,
		Kind:      Kind(r.Kind),
		TopicSlug: r.TopicSlug,
		Status:    r.Status,
		Error:     r.Error,
		CreatedAt: r.CreatedAt,
	}
	if r.ResultID.Valid {
		job.ResultID = &r.ResultID.Int64
	}
	if r.FinishedAt.Valid {
		job.FinishedAt = &r.FinishedAt.Int64
	}
	if r.Status == statusRunning {
		m.mu.Lock()
		if b, ok := m.partial[r.ID]; ok {
			job.Partial = b.String()
		}
		m.mu.Unlock()
	}
	return job
}
