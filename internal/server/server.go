package server

import (
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Module interface {
	Routes(mux *http.ServeMux)
}

type Authenticator interface {
	Module
	Require(next http.Handler) http.Handler
}

func Handler(auth Authenticator, assets fs.FS, modules ...Module) http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("GET /api/auth/me", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	for _, m := range modules {
		m.Routes(api)
	}

	root := http.NewServeMux()
	auth.Routes(root)
	root.Handle("GET /healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	root.Handle("/api/", auth.Require(api))
	root.Handle("/", spaHandler(assets))
	return logRequests(root)
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		if strings.HasPrefix(r.URL.Path, "/api/") {
			slog.Info("http", "method", r.Method, "path", r.URL.Path, "status", sw.status, "duration", time.Since(start).Round(time.Millisecond))
		}
	})
}
