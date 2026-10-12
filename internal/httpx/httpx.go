package httpx

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
)

type Error struct {
	Status  int
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Message }

func Errorf(status int, format string, args ...any) error {
	return &Error{Status: status, Message: fmt.Sprintf(format, args...)}
}

func Coded(status int, code, message string) error {
	return &Error{Status: status, Code: code, Message: message}
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func Handle(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			WriteError(w, r, err)
		}
	}
}

func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var httpErr *Error
	switch {
	case errors.As(err, &httpErr):
		body := map[string]string{"error": httpErr.Message}
		if httpErr.Code != "" {
			body["code"] = httpErr.Code
		}
		JSON(w, httpErr.Status, body)
	case errors.Is(err, sql.ErrNoRows):
		JSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	case r.Context().Err() != nil:
	default:
		slog.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
		JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
}

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		json.NewEncoder(w).Encode(v)
	}
}

func Decode(r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return Errorf(http.StatusBadRequest, "invalid request body: %v", err)
	}
	return nil
}

func PathID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil {
		return 0, Errorf(http.StatusBadRequest, "invalid %s", name)
	}
	return id, nil
}
