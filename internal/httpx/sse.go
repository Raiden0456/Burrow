package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type SSE struct {
	w  http.ResponseWriter
	rc *http.ResponseController
}

func NewSSE(w http.ResponseWriter) *SSE {
	rc := http.NewResponseController(w)
	rc.SetWriteDeadline(time.Time{})
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	return &SSE{w: w, rc: rc}
}

func (s *SSE) Send(event string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", event, payload); err != nil {
		return err
	}
	return s.rc.Flush()
}

func (s *SSE) Delta(text string) error {
	return s.Send("delta", map[string]string{"text": text})
}

func (s *SSE) Fail(err error) {
	body := map[string]string{"error": err.Error()}
	var httpErr *Error
	if errors.As(err, &httpErr) && httpErr.Code != "" {
		body["code"] = httpErr.Code
	}
	s.Send("error", body)
}
