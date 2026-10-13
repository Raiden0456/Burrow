package jobs

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/raiden0456/gonteractive/internal/store"
	"github.com/raiden0456/gonteractive/internal/store/db"
)

func newManager(t *testing.T) (*Manager, *db.Queries) {
	t.Helper()
	conn, err := store.Open(t.Context(), filepath.Join(t.TempDir(), "jobs.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	q := db.New(conn)
	m, err := New(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	return m, q
}

func waitFinished(t *testing.T, m *Manager, id int64) Job {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		job, err := m.Get(context.Background(), id)
		if err != nil {
			t.Fatal(err)
		}
		if job.Status != statusRunning {
			return job
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("job did not finish")
	return Job{}
}

func TestJobOutlivesRequestAndExposesPartial(t *testing.T) {
	m, _ := newManager(t)
	reqCtx, cancelReq := context.WithCancel(context.Background())
	release := make(chan struct{})

	job, err := m.Start(reqCtx, KindLesson, "maps", func(ctx context.Context, progress func(string)) (int64, error) {
		progress("hello ")
		progress("world")
		<-release
		return 42, ctx.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	cancelReq()

	deadline := time.Now().Add(5 * time.Second)
	for {
		running, err := m.Get(context.Background(), job.ID)
		if err != nil {
			t.Fatal(err)
		}
		if running.Partial == "hello world" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("partial never reached the expected text, got %q", running.Partial)
		}
		time.Sleep(10 * time.Millisecond)
	}

	again, err := m.Start(context.Background(), KindLesson, "maps", func(context.Context, func(string)) (int64, error) {
		t.Error("a second job for the same topic must not start")
		return 0, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != job.ID {
		t.Fatalf("expected the running job %d, got %d", job.ID, again.ID)
	}

	close(release)
	done := waitFinished(t, m, job.ID)
	if done.Status != statusDone || done.ResultID == nil || *done.ResultID != 42 || done.Partial != "" {
		t.Fatalf("unexpected finished job: %+v", done)
	}
}

func TestJobFailureIsRecorded(t *testing.T) {
	m, _ := newManager(t)
	job, err := m.Start(t.Context(), KindQuiz, "maps", func(context.Context, func(string)) (int64, error) {
		return 0, errors.New("model exploded")
	})
	if err != nil {
		t.Fatal(err)
	}
	done := waitFinished(t, m, job.ID)
	if done.Status != statusFailed || done.Error != "model exploded" {
		t.Fatalf("unexpected job: %+v", done)
	}
	recent, err := m.Recent(t.Context(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(recent) != 1 || recent[0].ID != job.ID {
		t.Fatalf("expected the failed job in recent jobs, got %+v", recent)
	}
}

func TestRestartFailsInterruptedJobs(t *testing.T) {
	_, q := newManager(t)
	row, err := q.CreateJob(t.Context(), db.CreateJobParams{Kind: string(KindDrill), TopicSlug: "maps"})
	if err != nil {
		t.Fatal(err)
	}
	m, err := New(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	job, err := m.Get(t.Context(), row.ID)
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != statusFailed {
		t.Fatalf("expected interrupted job to be failed, got %+v", job)
	}
}
