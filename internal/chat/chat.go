package chat

import (
	"net/http"
	"strings"

	"github.com/raiden0456/gonteractive/internal/ai"
	"github.com/raiden0456/gonteractive/internal/httpx"
	"github.com/raiden0456/gonteractive/internal/providers"
	"github.com/raiden0456/gonteractive/internal/settings"
	"github.com/raiden0456/gonteractive/internal/store/db"
	"github.com/raiden0456/gonteractive/internal/tutor"
)

const role = `You are the learner's personal Go mentor. Answer any question about Go, its ecosystem and software engineering in Go.
Be direct and concrete: lead with the answer, then show a minimal example when it helps. Point out common pitfalls when relevant.
If the learner asks you to quiz them or check their understanding, do it interactively one question at a time.`

type Service struct {
	q         *db.Queries
	providers *providers.Service
	settings  *settings.Store
}

func New(q *db.Queries, p *providers.Service, s *settings.Store) *Service {
	return &Service{q: q, providers: p, settings: s}
}

type threadView struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Context   string `json:"context"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type messageView struct {
	ID        int64  `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"createdAt"`
}

func (s *Service) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/chat/threads", httpx.Handle(s.listThreads))
	mux.HandleFunc("POST /api/chat/threads", httpx.Handle(s.createThread))
	mux.HandleFunc("GET /api/chat/threads/{id}", httpx.Handle(s.getThread))
	mux.HandleFunc("DELETE /api/chat/threads/{id}", httpx.Handle(s.deleteThread))
	mux.HandleFunc("POST /api/chat/threads/{id}/messages", httpx.Handle(s.send))
}

func (s *Service) listThreads(w http.ResponseWriter, r *http.Request) error {
	rows, err := s.q.ListThreads(r.Context(), 100)
	if err != nil {
		return err
	}
	out := make([]threadView, 0, len(rows))
	for _, t := range rows {
		out = append(out, toThreadView(t))
	}
	httpx.JSON(w, http.StatusOK, out)
	return nil
}

func (s *Service) createThread(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Title   string `json:"title"`
		Context string `json:"context"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	t, err := s.q.CreateThread(r.Context(), db.CreateThreadParams{Title: strings.TrimSpace(in.Title), Context: in.Context})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, toThreadView(t))
	return nil
}

func (s *Service) getThread(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	t, err := s.q.GetThread(r.Context(), id)
	if err != nil {
		return err
	}
	msgs, err := s.q.ListMessages(r.Context(), id)
	if err != nil {
		return err
	}
	views := make([]messageView, 0, len(msgs))
	for _, m := range msgs {
		views = append(views, toMessageView(m))
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"thread": toThreadView(t), "messages": views})
	return nil
}

func (s *Service) deleteThread(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	if err := s.q.DeleteThread(r.Context(), id); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Service) send(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in struct {
		Content string `json:"content"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	content := strings.TrimSpace(in.Content)
	if content == "" {
		return httpx.Errorf(http.StatusBadRequest, "message must not be empty")
	}

	thread, err := s.q.GetThread(ctx, id)
	if err != nil {
		return err
	}
	prefs, err := s.settings.Preferences(ctx)
	if err != nil {
		return err
	}
	provider, err := s.providers.For(ctx, providers.TaskChat)
	if err != nil {
		return err
	}
	history, err := s.q.ListMessages(ctx, id)
	if err != nil {
		return err
	}

	userMsg, err := s.q.CreateMessage(ctx, db.CreateMessageParams{ThreadID: id, Role: string(ai.RoleUser), Content: content})
	if err != nil {
		return err
	}
	if thread.Title == "" {
		thread.Title = titleFrom(content)
		if _, err := s.q.UpdateThreadTitle(ctx, db.UpdateThreadTitleParams{ID: id, Title: thread.Title}); err != nil {
			return err
		}
	}

	req := ai.Request{System: tutor.System(prefs, role)}
	if thread.Context != "" {
		req.System += "\n\nThe learner opened this conversation from the following context; use it to interpret their questions:\n<context>\n" + thread.Context + "\n</context>"
	}
	for _, m := range history {
		req.Messages = append(req.Messages, ai.Message{Role: ai.Role(m.Role), Content: m.Content})
	}
	req.Messages = append(req.Messages, ai.Message{Role: ai.RoleUser, Content: content})

	sse := httpx.NewSSE(w)
	sse.Send("user", toMessageView(userMsg))
	reply, err := provider.Stream(ctx, req, sse.Delta)
	if err != nil {
		sse.Fail(err)
		return nil
	}
	assistantMsg, err := s.q.CreateMessage(ctx, db.CreateMessageParams{ThreadID: id, Role: string(ai.RoleAssistant), Content: reply})
	if err != nil {
		sse.Fail(err)
		return nil
	}
	if err := s.q.TouchThread(ctx, id); err != nil {
		sse.Fail(err)
		return nil
	}
	sse.Send("done", map[string]any{"message": toMessageView(assistantMsg), "title": thread.Title})
	return nil
}

func titleFrom(content string) string {
	t := strings.Join(strings.Fields(content), " ")
	if r := []rune(t); len(r) > 60 {
		return string(r[:60]) + "…"
	}
	return t
}

func toThreadView(t db.ChatThread) threadView {
	return threadView{ID: t.ID, Title: t.Title, Context: t.Context, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt}
}

func toMessageView(m db.ChatMessage) messageView {
	return messageView{ID: m.ID, Role: m.Role, Content: m.Content, CreatedAt: m.CreatedAt}
}
