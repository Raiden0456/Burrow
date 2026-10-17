package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/raiden0456/gonteractive/internal/auth"
	"github.com/raiden0456/gonteractive/internal/chat"
	"github.com/raiden0456/gonteractive/internal/config"
	"github.com/raiden0456/gonteractive/internal/curriculum"
	"github.com/raiden0456/gonteractive/internal/daily"
	"github.com/raiden0456/gonteractive/internal/drill"
	"github.com/raiden0456/gonteractive/internal/jobs"
	"github.com/raiden0456/gonteractive/internal/lesson"
	"github.com/raiden0456/gonteractive/internal/progress"
	"github.com/raiden0456/gonteractive/internal/providers"
	"github.com/raiden0456/gonteractive/internal/quiz"
	"github.com/raiden0456/gonteractive/internal/server"
	"github.com/raiden0456/gonteractive/internal/settings"
	"github.com/raiden0456/gonteractive/internal/store"
	"github.com/raiden0456/gonteractive/internal/store/db"
	"github.com/raiden0456/gonteractive/internal/wallpaper"
	"github.com/raiden0456/gonteractive/web"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	conn, err := store.Open(ctx, cfg.DBPath())
	if err != nil {
		return err
	}
	defer conn.Close()
	q := db.New(conn)

	cur := curriculum.MustLoad()
	settingsStore := settings.New(q)
	providerService := providers.New(q, settingsStore)
	defer providerService.Close()
	progressService := progress.New(q, cur)
	jobsCtx, cancelJobs := context.WithCancel(ctx)
	jobManager, err := jobs.New(jobsCtx, q)
	if err != nil {
		cancelJobs()
		return err
	}
	defer jobManager.Wait()
	defer cancelJobs()
	quizService := quiz.New(conn, q, cur, providerService, settingsStore, progressService, jobManager)

	runner, runnerErr := drill.NewRunner(cfg.GoCacheDir())
	if runnerErr != nil {
		slog.Warn("drills disabled", "err", runnerErr)
	}

	pusher, err := daily.NewPusher(ctx, q, settingsStore, cfg.PushContact)
	if err != nil {
		return err
	}
	scheduler := daily.NewScheduler(q, settingsStore, quizService, pusher, jobManager)
	go scheduler.Run(ctx)

	handler := server.Handler(
		auth.New(q, cfg.Password, cfg.CookieSecure),
		web.Dist(),
		cur,
		settingsStore,
		wallpaper.New(cfg.DataDir),
		providerService,
		progressService,
		jobManager,
		lesson.New(q, cur, providerService, settingsStore, jobManager),
		quizService,
		drill.New(q, cur, providerService, settingsStore, jobManager, runner, runnerErr),
		chat.New(q, providerService, settingsStore),
		scheduler,
	)

	srv := &http.Server{Addr: cfg.Addr, Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	errc := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", cfg.Addr)
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
