package drill

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/raiden0456/gonteractive/internal/curriculum"
	"github.com/raiden0456/gonteractive/internal/jobs"
	"github.com/raiden0456/gonteractive/internal/providers"
	"github.com/raiden0456/gonteractive/internal/settings"
	"github.com/raiden0456/gonteractive/internal/store"
	"github.com/raiden0456/gonteractive/internal/store/db"
)

func bugHuntFixture() generated {
	return generated{
		Title:       "Unexpected double",
		Description: "Double should return twice its input, but Double(21) returns 21. Fix the implementation without changing its API.",
		StarterCode: "package drill\n\nfunc Double(n int) int { return n }\n",
		TestCode:    testFile,
		Solution:    "package drill\n\nfunc Double(n int) int { return n * 2 }\n",
		Explanation: "The original returned its input unchanged. Multiplying by two fixes the regression.",
		Difficulty:  1,
	}
}

func newTestService(t *testing.T) (*Service, *http.ServeMux) {
	t.Helper()
	conn, err := store.Open(t.Context(), filepath.Join(t.TempDir(), "drills.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	q := db.New(conn)
	prefs := settings.New(q)
	p := providers.New(q, prefs)
	t.Cleanup(func() { p.Close() })
	j, err := jobs.New(t.Context(), q)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(j.Wait)
	s := New(q, curriculum.MustLoad(), p, prefs, j, newTestRunner(t), nil)
	mux := http.NewServeMux()
	s.Routes(mux)
	return s, mux
}

func serve(t *testing.T, mux *http.ServeMux, method, path, body string, status int) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
	if w.Code != status {
		t.Fatalf("%s %s: status %d, want %d; body: %s", method, path, w.Code, status, w.Body.String())
	}
	return w
}

func TestBugHuntWorkflow(t *testing.T) {
	s, mux := newTestService(t)
	g := bugHuntFixture()
	var calls atomic.Int32
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if len(req.Messages) < 2 || !strings.Contains(req.Messages[1].Content, "BUG HUNT") {
			t.Error("generation request must use the Bug Hunt prompt")
		}
		calls.Add(1)
		content, _ := json.Marshal(g)
		chunk, _ := json.Marshal(map[string]any{
			"id": "test", "object": "chat.completion.chunk",
			"choices": []any{map[string]any{"index": 0, "delta": map[string]string{"content": string(content)}}},
		})
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: %s\n\ndata: [DONE]\n\n", chunk)
	}))
	defer fake.Close()
	cfg, _ := json.Marshal(providers.Config{Model: "test", BaseURL: fake.URL + "/v1", APIKey: "test"})
	if _, err := s.q.CreateProvider(t.Context(), db.CreateProviderParams{Name: "test", Kind: "openai", Config: string(cfg)}); err != nil {
		t.Fatal(err)
	}
	slug := s.curriculum.Ordered()[0]
	w := serve(t, mux, "POST", "/api/topics/"+slug+"/drills?mode=bug_hunt", "", http.StatusAccepted)
	var job jobs.Job
	if err := json.Unmarshal(w.Body.Bytes(), &job); err != nil {
		t.Fatal(err)
	}
	s.jobs.Wait()
	finished, err := s.jobs.Get(t.Context(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if finished.Status != "done" || finished.ResultID == nil {
		t.Fatalf("generation failed: %+v", finished)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected one provider call, got %d", calls.Load())
	}
	path := fmt.Sprintf("/api/drills/%d", *finished.ResultID)
	w = serve(t, mux, "GET", path, "", http.StatusOK)
	var view drillView
	if err := json.Unmarshal(w.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	if view.Mode != ModeBugHunt || view.TestCode != g.TestCode || !strings.Contains(view.InitialOutput, "want 42") {
		t.Fatalf("missing Bug Hunt context: %+v", view)
	}
	if view.LastAttempt != nil || view.Solved {
		t.Fatal("verification must not count as a learner attempt or solve the hunt")
	}
	if strings.Contains(w.Body.String(), g.Explanation) || strings.Contains(w.Body.String(), `return n * 2`) {
		t.Fatal("unsolved hunt leaked the explanation or solution")
	}
	serve(t, mux, "GET", path+"/solution", "", http.StatusForbidden)

	body, _ := json.Marshal(map[string]string{"code": g.StarterCode})
	serve(t, mux, "PUT", path+"/draft", string(body), http.StatusNoContent)
	serve(t, mux, "POST", path+"/run", string(body), http.StatusOK)
	serve(t, mux, "GET", path+"/solution", "", http.StatusForbidden)
	d, err := s.q.GetDrill(t.Context(), view.ID)
	if err != nil || d.Draft != g.StarterCode || d.SolvedAt.Valid {
		t.Fatalf("draft/failing attempt not persisted correctly: %+v, %v", d, err)
	}

	body, _ = json.Marshal(map[string]string{"code": g.Solution})
	serve(t, mux, "POST", path+"/run", string(body), http.StatusOK)
	w = serve(t, mux, "GET", path+"/solution", "", http.StatusOK)
	var solution map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &solution); err != nil {
		t.Fatal(err)
	}
	if solution["explanation"] != g.Explanation || solution["solution"] != g.Solution {
		t.Fatalf("missing debrief: %v", solution)
	}
	w = serve(t, mux, "GET", "/api/topics/"+slug+"/drills", "", http.StatusOK)
	var summaries []summaryView
	if err := json.Unmarshal(w.Body.Bytes(), &summaries); err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 1 || summaries[0].Mode != ModeBugHunt || !summaries[0].Solved {
		t.Fatalf("hunt should be listed as solved: %+v", summaries)
	}
}

func TestVerifyDrill(t *testing.T) {
	s := &Service{runner: newTestRunner(t)}
	tests := []struct {
		name    string
		change  func(*generated)
		problem string
	}{
		{"valid bug", func(*generated) {}, ""},
		{"passing starter", func(g *generated) { g.StarterCode = g.Solution }, "already passes"},
		{"broken starter", func(g *generated) { g.StarterCode = `package drill; func Double(n int) int { return "bad" }` }, "does not compile"},
		{"broken solution", func(g *generated) { g.Solution = g.StarterCode }, "reference solution does not pass"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := bugHuntFixture()
			tt.change(&g)
			initial, problem, err := s.verify(t.Context(), g)
			if err != nil {
				t.Fatal(err)
			}
			if tt.problem == "" {
				if problem != "" || initial.Passed || initial.BuildError || initial.TimedOut {
					t.Fatalf("expected verified failing starter, got %+v, %q", initial, problem)
				}
			} else if !strings.Contains(problem, tt.problem) {
				t.Fatalf("expected %q, got %q", tt.problem, problem)
			}
		})
	}
}

func TestDrillModeCompatibility(t *testing.T) {
	s, mux := newTestService(t)
	slug := s.curriculum.Ordered()[0]
	serve(t, mux, "POST", "/api/topics/"+slug+"/drills?mode=unknown", "", http.StatusBadRequest)
	serve(t, mux, "POST", "/api/topics/"+slug+"/drills", "", http.StatusPreconditionFailed)
	serve(t, mux, "POST", "/api/topics/"+slug+"/drills?mode=bug_hunt", "", http.StatusPreconditionFailed)
	g := bugHuntFixture()
	d, err := s.q.CreateDrill(t.Context(), db.CreateDrillParams{
		TopicSlug: slug, Title: g.Title, Description: g.Description,
		StarterCode: g.StarterCode, TestCode: g.TestCode, Solution: g.Solution,
		Difficulty: 1, Mode: string(ModeImplement),
	})
	if err != nil {
		t.Fatal(err)
	}
	w := serve(t, mux, "GET", fmt.Sprintf("/api/drills/%d/solution", d.ID), "", http.StatusOK)
	if !strings.Contains(w.Body.String(), "solution") {
		t.Fatal("regular drills must still allow revealing the solution before solving")
	}
	view := toView(d, nil)
	if view.TestCode != "" || view.InitialOutput != "" || view.Mode != ModeImplement {
		t.Fatalf("regular drill view changed unexpectedly: %+v", view)
	}
}

func TestGenerationPromptAndValidation(t *testing.T) {
	topic := curriculum.Topic{Title: "Slices", Objectives: []string{"Understand backing arrays"}}
	bug := generationPrompt(topic, []string{"Prior hunt"}, ModeBugHunt)
	regular := generationPrompt(topic, nil, ModeImplement)
	for _, want := range []string{"BUG HUNT", "No TODOs", "root cause", "Prior hunt", "Slices"} {
		if !strings.Contains(bug, want) {
			t.Fatalf("Bug Hunt prompt missing %q", want)
		}
	}
	if strings.Contains(bug, "Bodies contain a TODO") || !strings.Contains(regular, "Bodies contain a TODO") {
		t.Fatal("Bug Hunt and implementation instructions must remain distinct")
	}
	g := bugHuntFixture()
	if err := g.Validate(); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(g)
	if err := drillSchema.Validate(raw); err != nil {
		t.Fatal(err)
	}
	g.Explanation = " "
	if g.Validate() == nil {
		t.Fatal("an empty explanation must be rejected")
	}
}
