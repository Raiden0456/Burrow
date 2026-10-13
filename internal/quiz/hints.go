package quiz

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/raiden0456/gonteractive/internal/ai"
	"github.com/raiden0456/gonteractive/internal/httpx"
	"github.com/raiden0456/gonteractive/internal/providers"
	"github.com/raiden0456/gonteractive/internal/store/db"
	"github.com/raiden0456/gonteractive/internal/tutor"
)

const (
	hintKindHint    = "hint"
	hintKindExplain = "explain"
	maxHints        = 3
)

const hintRole = "You help a learner who is stuck on a quiz question, without giving the answer away."

const explainRole = "You explain quiz questions in depth after the learner has answered them."

type hintRequest struct {
	quiz     db.Quiz
	question db.Question
	answer   *db.Answer
	previous []db.Hint
}

func (s *Service) loadHintRequest(ctx context.Context, quizID, questionID int64) (hintRequest, error) {
	quiz, err := s.q.GetQuiz(ctx, quizID)
	if err != nil {
		return hintRequest{}, err
	}
	questions, err := s.q.ListQuizQuestions(ctx, quizID)
	if err != nil {
		return hintRequest{}, err
	}
	req := hintRequest{quiz: quiz}
	found := false
	for _, q := range questions {
		if q.ID == questionID {
			req.question, found = q, true
		}
	}
	if !found {
		return hintRequest{}, httpx.Errorf(http.StatusNotFound, "question %d is not part of quiz %d", questionID, quizID)
	}
	answers, err := s.q.ListQuizAnswers(ctx, sql.NullInt64{Int64: quizID, Valid: true})
	if err != nil {
		return hintRequest{}, err
	}
	for i := range answers {
		if answers[i].QuestionID == questionID {
			req.answer = &answers[i]
		}
	}
	req.previous, err = s.q.ListQuestionHints(ctx, db.ListQuestionHintsParams{
		QuizID:     sql.NullInt64{Int64: quizID, Valid: true},
		QuestionID: sql.NullInt64{Int64: questionID, Valid: true},
	})
	return req, err
}

// StreamHint generates a hint (before answering) or an in-depth explanation (after answering) for a
// quiz question, streams it through onDelta and stores it with the question.
func (s *Service) StreamHint(ctx context.Context, quizID, questionID int64, kind string, onDelta func(string) error) (Hint, error) {
	req, err := s.loadHintRequest(ctx, quizID, questionID)
	if err != nil {
		return Hint{}, err
	}

	var role, prompt string
	switch kind {
	case hintKindHint:
		if req.answer != nil {
			return Hint{}, httpx.Errorf(http.StatusConflict, "the question is already answered")
		}
		count := 0
		for _, h := range req.previous {
			if h.Kind == hintKindHint {
				count++
			}
		}
		if count >= maxHints {
			return Hint{}, httpx.Coded(http.StatusConflict, "hint_limit", "no more hints for this question")
		}
		role, prompt = hintRole, hintPrompt(req, count+1)
	case hintKindExplain:
		if req.answer == nil {
			return Hint{}, httpx.Errorf(http.StatusConflict, "answer the question first")
		}
		role, prompt = explainRole, explainPrompt(req)
	default:
		return Hint{}, httpx.Errorf(http.StatusBadRequest, "kind must be hint or explain")
	}

	prefs, err := s.settings.Preferences(ctx)
	if err != nil {
		return Hint{}, err
	}
	provider, err := s.providers.For(ctx, providers.TaskChat)
	if err != nil {
		return Hint{}, err
	}
	content, err := provider.Stream(ctx, ai.Request{
		System:   tutor.System(prefs, role),
		Messages: []ai.Message{{Role: ai.RoleUser, Content: prompt}},
	}, onDelta)
	if err != nil {
		return Hint{}, err
	}
	row, err := s.q.CreateHint(ctx, db.CreateHintParams{
		Kind:       kind,
		QuizID:     sql.NullInt64{Int64: quizID, Valid: true},
		QuestionID: sql.NullInt64{Int64: questionID, Valid: true},
		Content:    content,
	})
	if err != nil {
		return Hint{}, err
	}
	return toHint(row), nil
}

func questionBlock(q db.Question) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<question kind=%q>\n%s\n", q.Kind, q.Prompt)
	var options []string
	if json.Unmarshal([]byte(q.Options), &options) == nil && len(options) > 0 {
		b.WriteString("\nOptions:\n")
		for i, o := range options {
			fmt.Fprintf(&b, "%d. %s\n", i+1, o)
		}
	}
	b.WriteString("</question>\n")
	return b.String()
}

func hintPrompt(req hintRequest, level int) string {
	var b strings.Builder
	b.WriteString(questionBlock(req.question))
	for i, h := range req.previous {
		if h.Kind == hintKindHint {
			fmt.Fprintf(&b, "\n<previous_hint n=\"%d\">\n%s\n</previous_hint>\n", i+1, h.Content)
		}
	}
	fmt.Fprintf(&b, "\nWrite hint %d of %d for the learner.\n", level, maxHints)
	switch level {
	case 1:
		b.WriteString("Point at the concept or the part of the code to look at. Do not mention any option as right or wrong.")
	case 2:
		b.WriteString("Be more specific: explain the relevant rule of the language, possibly with a tiny unrelated example. Still do not name the answer.")
	default:
		b.WriteString("Walk through the reasoning almost to the end, so the learner only has to take the last step. Do not state the final answer.")
	}
	b.WriteString(" Keep it to 2-5 sentences. Do not repeat earlier hints.")
	return b.String()
}

func explainPrompt(req hintRequest) string {
	var b strings.Builder
	b.WriteString(questionBlock(req.question))
	var correct []int
	json.Unmarshal([]byte(req.question.Correct), &correct)
	if len(correct) > 0 {
		names := make([]string, len(correct))
		for i, c := range correct {
			names[i] = fmt.Sprint(c + 1)
		}
		fmt.Fprintf(&b, "\nCorrect option(s): %s\n", strings.Join(names, ", "))
	}
	if req.question.ReferenceAnswer != "" {
		fmt.Fprintf(&b, "\nReference answer: %s\n", req.question.ReferenceAnswer)
	}
	fmt.Fprintf(&b, "\nShort explanation already shown to the learner: %s\n", req.question.Explanation)
	fmt.Fprintf(&b, "\nThe learner answered: %s (score %.0f%%)\n", describeResponse(req.answer.Response), req.answer.Score*100)
	if req.answer.Feedback != "" {
		fmt.Fprintf(&b, "Grader feedback: %s\n", req.answer.Feedback)
	}
	b.WriteString("\nExplain the question in depth: why the right answer is right, what exactly is wrong or missing in the learner's answer, and the underlying rule of Go. Include one small code example. Do not repeat the short explanation verbatim.")
	return b.String()
}
