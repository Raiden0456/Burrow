package quiz

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/raiden0456/gonteractive/internal/ai"
	"github.com/raiden0456/gonteractive/internal/curriculum"
	"github.com/raiden0456/gonteractive/internal/providers"
	"github.com/raiden0456/gonteractive/internal/store/db"
	"github.com/raiden0456/gonteractive/internal/tutor"
)

var questionsSchema = ai.MustSchema("quiz_questions", []byte(`{
	"type": "object",
	"properties": {
		"questions": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
					"kind": {"type": "string", "enum": ["single_choice", "multi_choice", "open"]},
					"prompt": {"type": "string"},
					"options": {"type": "array", "items": {"type": "string"}},
					"correct": {"type": "array", "items": {"type": "integer"}},
					"referenceAnswer": {"type": "string"},
					"explanation": {"type": "string"},
					"difficulty": {"type": "integer", "enum": [1, 2, 3]}
				},
				"required": ["kind", "prompt", "options", "correct", "referenceAnswer", "explanation", "difficulty"],
				"additionalProperties": false
			}
		}
	},
	"required": ["questions"],
	"additionalProperties": false
}`))

const generatorRole = "You write quiz questions that check real understanding of Go, not trivia."

type generatedQuestion struct {
	Kind            Kind     `json:"kind"`
	Prompt          string   `json:"prompt"`
	Options         []string `json:"options"`
	Correct         []int    `json:"correct"`
	ReferenceAnswer string   `json:"referenceAnswer"`
	Explanation     string   `json:"explanation"`
	Difficulty      int      `json:"difficulty"`
}

type generatedSet struct {
	Questions []generatedQuestion `json:"questions"`
	want      int
}

func (g *generatedSet) Validate() error {
	if len(g.Questions) < g.want {
		return fmt.Errorf("expected %d questions, got %d", g.want, len(g.Questions))
	}
	g.Questions = g.Questions[:g.want]
	for i, q := range g.Questions {
		if err := q.validate(); err != nil {
			return fmt.Errorf("question %d: %w", i+1, err)
		}
	}
	return nil
}

func (q generatedQuestion) validate() error {
	if strings.TrimSpace(q.Prompt) == "" || strings.TrimSpace(q.Explanation) == "" {
		return fmt.Errorf("prompt and explanation must not be empty")
	}
	switch q.Kind {
	case KindOpen:
		if strings.TrimSpace(q.ReferenceAnswer) == "" {
			return fmt.Errorf("open questions need a referenceAnswer")
		}
		if len(q.Options) > 0 || len(q.Correct) > 0 {
			return fmt.Errorf("open questions must have empty options and correct")
		}
		return nil
	case KindSingle, KindMulti:
		if len(q.Options) < 3 || len(q.Options) > 6 {
			return fmt.Errorf("choice questions need 3-6 options, got %d", len(q.Options))
		}
		seen := map[int]bool{}
		for _, c := range q.Correct {
			if c < 0 || c >= len(q.Options) {
				return fmt.Errorf("correct index %d is out of range", c)
			}
			if seen[c] {
				return fmt.Errorf("correct index %d is duplicated", c)
			}
			seen[c] = true
		}
		if q.Kind == KindSingle && len(q.Correct) != 1 {
			return fmt.Errorf("single_choice needs exactly one correct index")
		}
		if q.Kind == KindMulti && len(q.Correct) < 2 {
			return fmt.Errorf("multi_choice needs at least two correct indices")
		}
		return nil
	default:
		return fmt.Errorf("unknown kind %q", q.Kind)
	}
}

func (s *Service) generateQuestions(ctx context.Context, topic curriculum.Topic, n int) ([]generatedQuestion, error) {
	prefs, err := s.settings.Preferences(ctx)
	if err != nil {
		return nil, err
	}
	provider, err := s.providers.For(ctx, providers.TaskGenerate)
	if err != nil {
		return nil, err
	}
	existing, err := s.q.ListRecentPromptsByTopic(ctx, db.ListRecentPromptsByTopicParams{TopicSlug: topic.Slug, Limit: 40})
	if err != nil {
		return nil, err
	}
	mistakes, err := s.q.ListRecentMistakesByTopic(ctx, db.ListRecentMistakesByTopicParams{TopicSlug: topic.Slug, Limit: 5})
	if err != nil {
		return nil, err
	}
	stats, err := s.progress.TopicStats(ctx)
	if err != nil {
		return nil, err
	}
	var mastery float64
	if st, ok := stats[topic.Slug]; ok {
		mastery = st.Mastery
	}

	set := generatedSet{want: n}
	err = ai.GenerateJSON(ctx, provider, ai.Request{
		System:   tutor.System(prefs, generatorRole),
		Messages: []ai.Message{{Role: ai.RoleUser, Content: generationPrompt(topic, n, mastery, existing, mistakes)}},
	}, questionsSchema, &set)
	if err != nil {
		return nil, err
	}
	return set.Questions, nil
}

func generationPrompt(topic curriculum.Topic, n int, mastery float64, existing []string, mistakes []db.ListRecentMistakesByTopicRow) string {
	var b strings.Builder
	b.WriteString(topic.Brief())
	fmt.Fprintf(&b, "\nCreate %d new quiz questions for this topic.\n", n)
	b.WriteString("Mix the kinds: mostly single_choice, some multi_choice, and about one open question per four questions.\n")

	switch {
	case mastery < 0.35:
		b.WriteString("The learner is new to this topic: use difficulty 1-2.\n")
	case mastery < 0.7:
		b.WriteString("The learner knows the basics of this topic: use difficulty 2, with a few 3s.\n")
	default:
		b.WriteString("The learner is strong on this topic: focus on difficulty 3 edge cases and subtle behaviour.\n")
	}

	b.WriteString(`
Rules:
- Prefer questions that make the learner reason: predict the output of a snippet, spot the bug, pick the idiomatic approach, explain behaviour.
- Code in prompts and options must be valid, realistic Go.
- single_choice: exactly one correct option. multi_choice: two or more correct options; the interface already tells the learner to select all that apply, so do not add that phrase to the prompt.
- Choice questions: 3-5 options with plausible distractors; "correct" holds zero-based option indices; "referenceAnswer" is "".
- open: the answer should take 1-5 sentences or a short snippet; "options" and "correct" are empty arrays; "referenceAnswer" is a model answer.
- "explanation" says why the right answer is right and, for choice questions, why the distractors are wrong.
- Do not reveal the answer in the prompt.
`)

	if len(existing) > 0 {
		b.WriteString("\nThe learner has already seen these questions; do not repeat them or ask the same thing in other words:\n")
		for _, p := range existing {
			fmt.Fprintf(&b, "- %s\n", truncate(p, 160))
		}
	}
	if len(mistakes) > 0 {
		b.WriteString("\nThe learner recently got these wrong; probe the same misconceptions from a different angle:\n")
		for _, m := range mistakes {
			fmt.Fprintf(&b, "- Question: %s\n  Their answer: %s\n", truncate(m.Prompt, 200), truncate(describeResponse(m.Response), 200))
		}
	}
	return b.String()
}

func describeResponse(raw string) string {
	var r Response
	if json.Unmarshal([]byte(raw), &r) != nil {
		return raw
	}
	if r.Text != "" {
		return r.Text
	}
	return fmt.Sprintf("selected options %v", r.Selected)
}

func truncate(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}
