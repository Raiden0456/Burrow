package quiz

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/raiden0456/gonteractive/internal/ai"
	"github.com/raiden0456/gonteractive/internal/providers"
	"github.com/raiden0456/gonteractive/internal/store/db"
	"github.com/raiden0456/gonteractive/internal/tutor"
)

var gradeSchema = ai.MustSchema("grade", []byte(`{
	"type": "object",
	"properties": {
		"score": {"type": "number"},
		"feedback": {"type": "string"}
	},
	"required": ["score", "feedback"],
	"additionalProperties": false
}`))

const graderRole = "You grade a learner's answers to open questions fairly and explain your verdict."

type grade struct {
	Score    float64 `json:"score"`
	Feedback string  `json:"feedback"`
}

func (g *grade) Validate() error {
	if g.Score < 0 || g.Score > 1 {
		return fmt.Errorf("score must be between 0 and 1, got %v", g.Score)
	}
	if strings.TrimSpace(g.Feedback) == "" {
		return fmt.Errorf("feedback must not be empty")
	}
	return nil
}

func gradeChoice(q db.Question, selected []int) (float64, error) {
	var correct []int
	if err := json.Unmarshal([]byte(q.Correct), &correct); err != nil {
		return 0, err
	}
	picked := map[int]bool{}
	for _, s := range selected {
		picked[s] = true
	}

	if Kind(q.Kind) == KindSingle {
		if len(picked) == 1 && picked[correct[0]] {
			return 1, nil
		}
		return 0, nil
	}

	var hits, misses int
	for s := range picked {
		if slices.Contains(correct, s) {
			hits++
		} else {
			misses++
		}
	}
	return max(0, float64(hits-misses)/float64(len(correct))), nil
}

func (s *Service) gradeOpen(ctx context.Context, q db.Question, answer string) (grade, error) {
	if strings.TrimSpace(answer) == "" {
		return grade{Score: 0, Feedback: "No answer given."}, nil
	}
	prefs, err := s.settings.Preferences(ctx)
	if err != nil {
		return grade{}, err
	}
	provider, err := s.providers.For(ctx, providers.TaskGrade)
	if err != nil {
		return grade{}, err
	}

	prompt := fmt.Sprintf(`Grade the learner's answer to this Go question.

<question>
%s
</question>

<reference_answer>
%s
</reference_answer>

<explanation>
%s
</explanation>

<learner_answer>
%s
</learner_answer>

Set "score" between 0 and 1: 0 for wrong or irrelevant, around 0.5 for partially correct with important gaps, 1 for fully correct. Different wording than the reference is fine; judge the substance.
Set "feedback" to 2-4 sentences addressed to the learner: what is right, what is missing or wrong, and the key idea to remember.`,
		q.Prompt, q.ReferenceAnswer, q.Explanation, answer)

	var g grade
	err = ai.GenerateJSON(ctx, provider, ai.Request{
		System:   tutor.System(prefs, graderRole),
		Messages: []ai.Message{{Role: ai.RoleUser, Content: prompt}},
	}, gradeSchema, &g)
	return g, err
}
