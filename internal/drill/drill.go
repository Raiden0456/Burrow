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
		"difficulty": {"type": "integer", "enum": [1, 2, 3]}
	},
	"required": ["title", "description", "starterCode", "testCode", "solution", "difficulty"],
	"additionalProperties": false
}`))

const generatorRole = "You design small, focused coding exercises with automated tests."

type generated struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	StarterCode string `json:"starterCode"`
	TestCode    string `json:"testCode"`
	Solution    string `json:"solution"`
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
	return nil
}

type Service struct {
	q          *db.Queries
	curriculum *curriculum.Curriculum
	providers  *providers.Service
	settings   *settings.Store
	runner     *Runner
	runnerErr  error
}

func New(q *db.Queries, c *curriculum.Curriculum, p *providers.Service, s *settings.Store, runner *Runner, runnerErr error) *Service {
	return &Service{q: q, curriculum: c, providers: p, settings: s, runner: runner, runnerErr: runnerErr}
}

// Generate asks the model for an exercise and only accepts it once the reference solution passes the
// tests and the starter code compiles against them; otherwise the model gets the failing output and
// one more attempt.
func (s *Service) Generate(ctx context.Context, slug string) (db.Drill, error) {
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
		Messages: []ai.Message{{Role: ai.RoleUser, Content: generationPrompt(topic, titles)}},
	}
	var lastErr error
	for range 2 {
		var g generated
		if err := ai.GenerateJSON(ctx, provider, req, drillSchema, &g); err != nil {
			return db.Drill{}, err
		}
		problem, err := s.verify(ctx, g)
		if err != nil {
			return db.Drill{}, err
		}
		if problem == "" {
			return s.q.CreateDrill(ctx, db.CreateDrillParams{
				TopicSlug:   slug,
				Title:       g.Title,
				Description: g.Description,
				StarterCode: g.StarterCode,
				TestCode:    g.TestCode,
				Solution:    g.Solution,
				Difficulty:  int64(g.Difficulty),
			})
		}
		lastErr = errors.New(problem)
		req.Messages = append(req.Messages,
			ai.Message{Role: ai.RoleAssistant, Content: fmt.Sprintf("Title: %s\n\nSolution:\n%s\n\nTests:\n%s", g.Title, g.Solution, g.TestCode)},
			ai.Message{Role: ai.RoleUser, Content: problem + "\nFix the exercise and return the complete corrected JSON object."},
		)
	}
	return db.Drill{}, fmt.Errorf("could not generate a working drill: %w", lastErr)
}

func (s *Service) verify(ctx context.Context, g generated) (string, error) {
	res, err := s.runner.Run(ctx, g.Solution, g.TestCode)
	if err != nil {
		return "", err
	}
	if !res.Passed {
		return "The reference solution does not pass the tests. `go test` output:\n" + res.Output, nil
	}
	res, err = s.runner.Run(ctx, g.StarterCode, g.TestCode)
	if err != nil {
		return "", err
	}
	if res.BuildError {
		return "The starter code does not compile together with the tests. `go test` output:\n" + res.Output, nil
	}
	if res.Passed {
		return "The starter code already passes the tests; it must leave the work to the learner.", nil
	}
	return "", nil
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

func generationPrompt(topic curriculum.Topic, existingTitles []string) string {
	var b strings.Builder
	b.WriteString(topic.Brief())
	b.WriteString(`
Design one coding exercise (a "drill") that practises this topic.

Rules:
- Everything is in a single package named drill, using only the standard library. No main function, no network, no files, no sleeping longer than a few milliseconds.
- "starterCode": a complete Go file (package drill) with the types and function signatures the learner must implement. Bodies contain a TODO comment and return zero values (or panic("TODO")) so the file compiles but the tests fail.
- "testCode": a complete _test.go file (package drill) with table-driven tests covering normal cases and edge cases. Use only the testing package and the standard library. Test failure messages must show input, got and want.
- "solution": a complete, idiomatic Go file (package drill) that passes every test.
- "description": a Markdown task statement for the learner: what to implement, the signature(s), constraints and one or two examples. Do not include the solution.
- Scope it so it takes 5-20 minutes. Concurrency drills must be deterministic and race-free.
`)
	if len(existingTitles) > 0 {
		b.WriteString("\nThe learner already has these drills for the topic; make something different:\n")
		for _, t := range existingTitles {
			fmt.Fprintf(&b, "- %s\n", t)
		}
	}
	return b.String()
}
