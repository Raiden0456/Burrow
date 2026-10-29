package drill

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"strings"

	"github.com/raiden0456/gonteractive/internal/ai"
	"github.com/raiden0456/gonteractive/internal/curriculum"
	"github.com/raiden0456/gonteractive/internal/jobs"
	"github.com/raiden0456/gonteractive/internal/providers"
	"github.com/raiden0456/gonteractive/internal/settings"
	"github.com/raiden0456/gonteractive/internal/store/db"
	"github.com/raiden0456/gonteractive/internal/tutor"
)

var drillSchema = ai.MustSchema("drill", []byte(`{
	"type": "object",
	"properties": {
		"title": {"type": "string"},
		"description": {"type": "string"},
		"starterCode": {"type": "string"},
		"testCode": {"type": "string"},
		"solution": {"type": "string"},
		"explanation": {"type": "string"},
		"difficulty": {"type": "integer", "enum": [1, 2, 3]}
	},
	"required": ["title", "description", "starterCode", "testCode", "solution", "explanation", "difficulty"],
	"additionalProperties": false
}`))

const generatorRole = "You design small, focused coding exercises with automated tests."

type Mode string

const (
	ModeImplement Mode = "implement"
	ModeBugHunt   Mode = "bug_hunt"
)

type generated struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	StarterCode string `json:"starterCode"`
	TestCode    string `json:"testCode"`
	Solution    string `json:"solution"`
	Explanation string `json:"explanation"`
	Difficulty  int    `json:"difficulty"`
}

func (g *generated) Validate() error {
	for name, src := range map[string]string{"starterCode": g.StarterCode, "testCode": g.TestCode, "solution": g.Solution} {
		f, err := parser.ParseFile(token.NewFileSet(), "", src, parser.PackageClauseOnly)
		if err != nil || f.Name.Name != "drill" {
			return fmt.Errorf("%s must be a Go file in package drill", name)
		}
	}
	if !strings.Contains(g.TestCode, "func Test") {
		return errors.New("testCode must contain at least one Test function")
	}
	if strings.TrimSpace(g.Title) == "" || strings.TrimSpace(g.Description) == "" {
		return errors.New("title and description must not be empty")
	}
	if strings.TrimSpace(g.Explanation) == "" {
		return errors.New("explanation must not be empty")
	}
	return nil
}

type Service struct {
	q          *db.Queries
	curriculum *curriculum.Curriculum
	providers  *providers.Service
	settings   *settings.Store
	jobs       *jobs.Manager
	runner     *Runner
	runnerErr  error
}

func New(q *db.Queries, c *curriculum.Curriculum, p *providers.Service, s *settings.Store, j *jobs.Manager, runner *Runner, runnerErr error) *Service {
	return &Service{q: q, curriculum: c, providers: p, settings: s, jobs: j, runner: runner, runnerErr: runnerErr}
}

// Generate asks the model for an exercise and only accepts it once the reference solution passes the
// tests and the starter code compiles against them; otherwise the model gets the failing output and
// one more attempt.
func (s *Service) Generate(ctx context.Context, slug string, mode Mode) (db.Drill, error) {
	if mode != ModeImplement && mode != ModeBugHunt {
		return db.Drill{}, fmt.Errorf("invalid drill mode: %q", mode)
	}
	if s.runner == nil {
		return db.Drill{}, s.runnerErr
	}
	topic, err := s.curriculum.Lookup(slug)
	if err != nil {
		return db.Drill{}, err
	}
	prefs, err := s.settings.Preferences(ctx)
	if err != nil {
		return db.Drill{}, err
	}
	provider, err := s.providers.For(ctx, providers.TaskGenerate)
	if err != nil {
		return db.Drill{}, err
	}
	titles, err := s.q.ListDrillTitlesByTopic(ctx, db.ListDrillTitlesByTopicParams{TopicSlug: slug, Limit: 30})
	if err != nil {
		return db.Drill{}, err
	}

	req := ai.Request{
		System:   tutor.System(prefs, generatorRole),
		Messages: []ai.Message{{Role: ai.RoleUser, Content: generationPrompt(topic, titles, mode)}},
	}
	var lastErr error
	for range 2 {
		var g generated
		if err := ai.GenerateJSON(ctx, provider, req, drillSchema, &g); err != nil {
			return db.Drill{}, err
		}
		initial, problem, err := s.verify(ctx, g)
		if err != nil {
			return db.Drill{}, err
		}
		if problem == "" {
			return s.q.CreateDrill(ctx, db.CreateDrillParams{
				TopicSlug:     slug,
				Title:         g.Title,
				Description:   g.Description,
				StarterCode:   g.StarterCode,
				TestCode:      g.TestCode,
				Solution:      g.Solution,
				Difficulty:    int64(g.Difficulty),
				Mode:          string(mode),
				Explanation:   g.Explanation,
				InitialOutput: initial.Output,
			})
		}
		lastErr = errors.New(problem)
		req.Messages = append(req.Messages,
			ai.Message{Role: ai.RoleAssistant, Content: fmt.Sprintf("Title: %s\n\nStarter:\n%s\n\nSolution:\n%s\n\nTests:\n%s", g.Title, g.StarterCode, g.Solution, g.TestCode)},
			ai.Message{Role: ai.RoleUser, Content: problem + "\nFix the exercise and return the complete corrected JSON object."},
		)
	}
	return db.Drill{}, fmt.Errorf("could not generate a working drill: %w", lastErr)
}

func (s *Service) verify(ctx context.Context, g generated) (RunResult, string, error) {
	res, err := s.runner.Run(ctx, g.Solution, g.TestCode)
	if err != nil {
		return res, "", err
	}
	if !res.Passed {
		return res, "The reference solution does not pass the tests. `go test` output:\n" + res.Output, nil
	}
	res, err = s.runner.Run(ctx, g.StarterCode, g.TestCode)
	if err != nil {
		return res, "", err
	}
	if res.BuildError {
		return res, "The starter code does not compile together with the tests. `go test` output:\n" + res.Output, nil
	}
	if res.TimedOut {
		return res, "The starter code times out. It must fail promptly and deterministically, not hang.", nil
	}
	if res.Passed {
		return res, "The starter code already passes the tests; it must leave the work to the learner.", nil
	}
	return res, "", nil
}

func (s *Service) Run(ctx context.Context, id int64, code string) (RunResult, error) {
	if s.runner == nil {
		return RunResult{}, s.runnerErr
	}
	d, err := s.q.GetDrill(ctx, id)
	if err != nil {
		return RunResult{}, err
	}
	res, err := s.runner.Run(ctx, code, d.TestCode)
	if err != nil {
		return RunResult{}, err
	}
	passed := int64(0)
	if res.Passed {
		passed = 1
	}
	if _, err := s.q.CreateDrillAttempt(ctx, db.CreateDrillAttemptParams{DrillID: id, Code: code, Passed: passed, Output: res.Output}); err != nil {
		return RunResult{}, err
	}
	if res.Passed {
		if err := s.q.MarkDrillSolved(ctx, id); err != nil {
			return RunResult{}, err
		}
	}
	return res, nil
}

func (s *Service) latestAttempt(ctx context.Context, id int64) (*db.DrillAttempt, error) {
	a, err := s.q.GetLatestDrillAttempt(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func generationPrompt(topic curriculum.Topic, existingTitles []string, mode Mode) string {
	var b strings.Builder
	b.WriteString(topic.Brief())
	b.WriteString(`
Design one coding exercise (a "drill") that practises this topic.

Rules:
- Everything is in a single package named drill, using only the standard library. No main function, no network, no files, no sleeping longer than a few milliseconds.
- "testCode": a complete _test.go file (package drill) with table-driven tests covering normal cases and edge cases. Use only the testing package and the standard library. Test failure messages must show input, got and want.
- "solution": a complete, idiomatic Go file (package drill) that passes every test.
- Scope it so it takes 5-20 minutes. Concurrency drills must be deterministic and race-free.
`)
	if mode == ModeBugHunt {
		b.WriteString(`
This is a BUG HUNT: the learner diagnoses and fixes existing code, not an implementation exercise.
- "starterCode": a complete, plausible implementation containing one focused, realistic Go bug related to this topic. No TODOs, placeholder bodies, deliberate panic("TODO"), or comments identifying the bug. It MUST compile against the tests and fail at least one test promptly and deterministically. No hangs, flaky timing, or data races.
- "description": a Markdown bug report describing expected behavior, observed symptoms, function signatures, and examples. Ask the learner to diagnose and fix the code while preserving its API. Do not reveal the root cause, location of the bug, or fix in the title or description.
- "testCode": regression tests that expose the bug and protect correct behavior. These tests and their initial failing output will be visible to the learner.
- "solution": the corrected implementation, preserving the starter's API.
- "explanation": a post-fix Markdown debrief explaining the root cause, why the original code failed, how the fix works, and a reusable Go debugging takeaway. It is hidden until the learner passes the tests.
`)
	} else {
		b.WriteString(`
- "starterCode": a complete Go file (package drill) with the types and function signatures the learner must implement. Bodies contain a TODO comment and return zero values (or panic("TODO")) so the file compiles but the tests fail.
- "description": a Markdown task statement for the learner: what to implement, the signature(s), constraints and one or two examples. Do not include the solution.
- "explanation": a short Markdown explanation of the reference solution.
`)
	}
	if len(existingTitles) > 0 {
		b.WriteString("\nThe learner already has these drills for the topic; make something different:\n")
		for _, t := range existingTitles {
			fmt.Fprintf(&b, "- %s\n", t)
		}
	}
	return b.String()
}
