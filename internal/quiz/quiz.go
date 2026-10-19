package quiz

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/raiden0456/gonteractive/internal/curriculum"
	"github.com/raiden0456/gonteractive/internal/httpx"
	"github.com/raiden0456/gonteractive/internal/jobs"
	"github.com/raiden0456/gonteractive/internal/progress"
	"github.com/raiden0456/gonteractive/internal/providers"
	"github.com/raiden0456/gonteractive/internal/settings"
	"github.com/raiden0456/gonteractive/internal/store"
	"github.com/raiden0456/gonteractive/internal/store/db"
)

type Kind string

const (
	KindSingle Kind = "single_choice"
	KindMulti  Kind = "multi_choice"
	KindOpen   Kind = "open"
)

const (
	quizTopic      = "topic"
	quizReview     = "review"
	quizDaily      = "daily"
	quizCheckpoint = "checkpoint"

	CheckpointSize = 8

	dailyReviewShare = 0.3
	dailyWeakShare   = 0.2
)

type Response struct {
	Selected []int  `json:"selected,omitempty"`
	Text     string `json:"text,omitempty"`
}

type Result struct {
	Response        Response `json:"response"`
	Score           float64  `json:"score"`
	Correct         bool     `json:"correct"`
	Feedback        string   `json:"feedback"`
	Explanation     string   `json:"explanation"`
	CorrectOptions  []int    `json:"correctOptions"`
	ReferenceAnswer string   `json:"referenceAnswer"`
}

type Hint struct {
	ID        int64  `json:"id"`
	Kind      string `json:"kind"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"createdAt"`
}

type QuestionView struct {
	ID         int64     `json:"id"`
	TopicSlug  string    `json:"topicSlug"`
	Kind       Kind      `json:"kind"`
	Prompt     string    `json:"prompt"`
	Options    []string  `json:"options"`
	Difficulty int64     `json:"difficulty"`
	Result     *Result   `json:"result"`
	Draft      *Response `json:"draft"`
	Hints      []Hint    `json:"hints"`
}

type QuizView struct {
	ID          int64          `json:"id"`
	Kind        string         `json:"kind"`
	TopicSlug   *string        `json:"topicSlug"`
	DailyDate   *string        `json:"dailyDate"`
	CreatedAt   int64          `json:"createdAt"`
	CompletedAt *int64         `json:"completedAt"`
	Score       *float64       `json:"score"`
	Questions   []QuestionView `json:"questions"`
}

type Summary struct {
	ID          int64    `json:"id"`
	Kind        string   `json:"kind"`
	TopicSlug   *string  `json:"topicSlug"`
	DailyDate   *string  `json:"dailyDate"`
	CreatedAt   int64    `json:"createdAt"`
	CompletedAt *int64   `json:"completedAt"`
	Total       int64    `json:"total"`
	Answered    int      `json:"answered"`
	Score       *float64 `json:"score"`
}

type Service struct {
	jobs       *jobs.Manager
	conn       *sql.DB
	q          *db.Queries
	curriculum *curriculum.Curriculum
	providers  *providers.Service
	settings   *settings.Store
	progress   *progress.Service
}

func New(conn *sql.DB, q *db.Queries, c *curriculum.Curriculum, p *providers.Service, s *settings.Store, prog *progress.Service, j *jobs.Manager) *Service {
	return &Service{conn: conn, q: q, curriculum: c, providers: p, settings: s, progress: prog, jobs: j}
}

func (s *Service) CreateTopicQuiz(ctx context.Context, slug string, size int) (db.Quiz, error) {
	topic, err := s.curriculum.Lookup(slug)
	if err != nil {
		return db.Quiz{}, err
	}
	generated, err := s.generateQuestions(ctx, topic, size, false)
	if err != nil {
		return db.Quiz{}, err
	}
	return s.createQuiz(ctx, db.CreateQuizParams{Kind: quizTopic, TopicSlug: nullString(slug)}, nil, map[string][]generatedQuestion{slug: generated})
}

func (s *Service) CreateCheckpointQuiz(ctx context.Context, slug string) (db.Quiz, error) {
	topic, err := s.curriculum.Lookup(slug)
	if err != nil {
		return db.Quiz{}, err
	}
	generated, err := s.generateQuestions(ctx, topic, CheckpointSize, true)
	if err != nil {
		return db.Quiz{}, err
	}
	return s.createQuiz(ctx, db.CreateQuizParams{Kind: quizTopic, TopicSlug: nullString(slug), Checkpoint: true}, nil, map[string][]generatedQuestion{slug: generated})
}

func (s *Service) CreateReviewQuiz(ctx context.Context, size int) (db.Quiz, error) {
	due, err := s.q.ListDueQuestionIDs(ctx, db.ListDueQuestionIDsParams{Due: time.Now().Unix(), Limit: int64(size)})
	if err != nil {
		return db.Quiz{}, err
	}
	if len(due) == 0 {
		return db.Quiz{}, httpx.Coded(http.StatusConflict, "nothing_due", "nothing is due for review")
	}
	return s.createQuiz(ctx, db.CreateQuizParams{Kind: quizReview}, due, nil)
}

// CreateDailyQuiz mixes due reviews (up to 30%), the weakest practiced topic (20%) and the next
// unlearned topic of the curriculum (the rest). It is idempotent per date; generation failures are
// tolerated as long as at least one question ends up in the quiz.
func (s *Service) CreateDailyQuiz(ctx context.Context, date string, size int) (db.Quiz, error) {
	if existing, err := s.q.GetDailyQuiz(ctx, nullString(date)); err == nil {
		return existing, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return db.Quiz{}, err
	}

	now := time.Now()
	reviewCap := int(math.Round(float64(size) * dailyReviewShare))
	due, err := s.q.ListDueQuestionIDs(ctx, db.ListDueQuestionIDsParams{Due: now.Unix(), Limit: int64(reviewCap)})
	if err != nil {
		return db.Quiz{}, err
	}
	frontier, weakest, err := s.progress.DailyTopics(ctx, now)
	if err != nil {
		return db.Quiz{}, err
	}

	plan := dailyPlan(size-len(due), int(math.Round(float64(size)*dailyWeakShare)), frontier, weakest)
	if len(plan) == 0 && len(due) < size {
		more, err := s.q.ListDueQuestionIDs(ctx, db.ListDueQuestionIDsParams{Due: now.Unix(), Limit: int64(size)})
		if err != nil {
			return db.Quiz{}, err
		}
		due = more
	}

	results := make([][]generatedQuestion, len(plan))
	errs := make([]error, len(plan))
	var wg sync.WaitGroup
	for i, p := range plan {
		topic, _ := s.curriculum.Topic(p.slug)
		wg.Go(func() { results[i], errs[i] = s.generateQuestions(ctx, topic, p.n, false) })
	}
	wg.Wait()

	fresh := map[string][]generatedQuestion{}
	var genErr error
	for i, p := range plan {
		if errors.Is(errs[i], providers.ErrNotConfigured) {
			return db.Quiz{}, errs[i]
		}
		if errs[i] != nil {
			genErr = errs[i]
			slog.Warn("daily quiz: question generation failed", "topic", p.slug, "err", errs[i])
			continue
		}
		fresh[p.slug] = results[i]
	}
	if len(due) == 0 && len(fresh) == 0 {
		return db.Quiz{}, fmt.Errorf("daily quiz: no due reviews and generation failed: %w", genErr)
	}
	return s.createQuiz(ctx, db.CreateQuizParams{Kind: quizDaily, DailyDate: nullString(date)}, due, fresh)
}

type topicShare struct {
	slug string
	n    int
}

func dailyPlan(fresh, weakShare int, frontier, weakest string) []topicShare {
	if fresh <= 0 {
		return nil
	}
	weakN := 0
	switch {
	case weakest == "":
	case frontier == "":
		weakN = fresh
	default:
		weakN = min(weakShare, fresh-1)
	}
	var plan []topicShare
	if frontier != "" {
		plan = append(plan, topicShare{frontier, fresh - weakN})
	}
	if weakN > 0 {
		plan = append(plan, topicShare{weakest, weakN})
	}
	return plan
}

func (s *Service) createQuiz(ctx context.Context, params db.CreateQuizParams, existing []int64, fresh map[string][]generatedQuestion) (db.Quiz, error) {
	var quiz db.Quiz
	err := store.Tx(ctx, s.conn, func(tx *sql.Tx) error {
		q := s.q.WithTx(tx)
		var err error
		if quiz, err = q.CreateQuiz(ctx, params); err != nil {
			return err
		}
		ids := existing
		for slug, gqs := range fresh {
			for _, gq := range gqs {
				options, _ := json.Marshal(gq.Options)
				correct, _ := json.Marshal(gq.Correct)
				row, err := q.CreateQuestion(ctx, db.CreateQuestionParams{
					TopicSlug:       slug,
					Kind:            string(gq.Kind),
					Prompt:          gq.Prompt,
					Options:         string(options),
					Correct:         string(correct),
					ReferenceAnswer: gq.ReferenceAnswer,
					Explanation:     gq.Explanation,
					Difficulty:      int64(gq.Difficulty),
				})
				if err != nil {
					return err
				}
				ids = append(ids, row.ID)
			}
		}
		for i, id := range ids {
			if err := q.AddQuizItem(ctx, db.AddQuizItemParams{QuizID: quiz.ID, Position: int64(i), QuestionID: id}); err != nil {
				return err
			}
		}
		return nil
	})
	return quiz, err
}

func (s *Service) View(ctx context.Context, id int64) (QuizView, error) {
	quiz, err := s.q.GetQuiz(ctx, id)
	if err != nil {
		return QuizView{}, err
	}
	questions, err := s.q.ListQuizQuestions(ctx, id)
	if err != nil {
		return QuizView{}, err
	}
	answers, err := s.q.ListQuizAnswers(ctx, sql.NullInt64{Int64: id, Valid: true})
	if err != nil {
		return QuizView{}, err
	}
	byQuestion := map[int64]db.Answer{}
	for _, a := range answers {
		byQuestion[a.QuestionID] = a
	}
	drafts, err := s.q.ListQuizDrafts(ctx, id)
	if err != nil {
		return QuizView{}, err
	}
	draftByQuestion := map[int64]string{}
	for _, d := range drafts {
		draftByQuestion[d.QuestionID] = d.Draft
	}
	hints, err := s.q.ListQuizHints(ctx, sql.NullInt64{Int64: id, Valid: true})
	if err != nil {
		return QuizView{}, err
	}
	hintsByQuestion := map[int64][]Hint{}
	for _, h := range hints {
		hintsByQuestion[h.QuestionID.Int64] = append(hintsByQuestion[h.QuestionID.Int64], toHint(h))
	}

	view := QuizView{
		ID:          quiz.ID,
		Kind:        kindOf(quiz),
		TopicSlug:   ptrString(quiz.TopicSlug),
		DailyDate:   ptrString(quiz.DailyDate),
		CreatedAt:   quiz.CreatedAt,
		CompletedAt: ptrInt(quiz.CompletedAt),
		Score:       averageScore(answers),
	}
	for _, q := range questions {
		qv := QuestionView{ID: q.ID, TopicSlug: q.TopicSlug, Kind: Kind(q.Kind), Prompt: q.Prompt, Difficulty: q.Difficulty, Hints: hintsByQuestion[q.ID]}
		if qv.Hints == nil {
			qv.Hints = []Hint{}
		}
		if err := json.Unmarshal([]byte(q.Options), &qv.Options); err != nil {
			return QuizView{}, err
		}
		if a, ok := byQuestion[q.ID]; ok {
			res, err := result(q, a)
			if err != nil {
				return QuizView{}, err
			}
			qv.Result = &res
		} else if raw, ok := draftByQuestion[q.ID]; ok {
			var draft Response
			if json.Unmarshal([]byte(raw), &draft) == nil {
				qv.Draft = &draft
			}
		}
		view.Questions = append(view.Questions, qv)
	}
	return view, nil
}

func (s *Service) Answer(ctx context.Context, quizID, questionID int64, resp Response) (Result, bool, error) {
	quiz, err := s.q.GetQuiz(ctx, quizID)
	if err != nil {
		return Result{}, false, err
	}
	questions, err := s.q.ListQuizQuestions(ctx, quizID)
	if err != nil {
		return Result{}, false, err
	}
	var question *db.Question
	for i := range questions {
		if questions[i].ID == questionID {
			question = &questions[i]
		}
	}
	if question == nil {
		return Result{}, false, httpx.Errorf(http.StatusBadRequest, "question %d is not part of quiz %d", questionID, quizID)
	}
	answers, err := s.q.ListQuizAnswers(ctx, sql.NullInt64{Int64: quizID, Valid: true})
	if err != nil {
		return Result{}, false, err
	}
	for _, a := range answers {
		if a.QuestionID == questionID {
			return Result{}, false, httpx.Coded(http.StatusConflict, "already_answered", "question already answered")
		}
	}

	var (
		score    float64
		feedback string
	)
	if Kind(question.Kind) == KindOpen {
		resp.Selected = nil
		g, err := s.gradeOpen(ctx, *question, resp.Text)
		if err != nil {
			return Result{}, false, err
		}
		score, feedback = g.Score, g.Feedback
	} else {
		resp.Text = ""
		if score, err = gradeChoice(*question, resp.Selected); err != nil {
			return Result{}, false, err
		}
	}

	hints, err := s.q.ListQuestionHints(ctx, db.ListQuestionHintsParams{
		QuizID:     sql.NullInt64{Int64: quizID, Valid: true},
		QuestionID: sql.NullInt64{Int64: questionID, Valid: true},
	})
	if err != nil {
		return Result{}, false, err
	}
	hinted := false
	for _, h := range hints {
		hinted = hinted || h.Kind == hintKindHint
	}

	rawResp, _ := json.Marshal(resp)
	var answer db.Answer
	completed := false
	err = store.Tx(ctx, s.conn, func(tx *sql.Tx) error {
		q := s.q.WithTx(tx)
		var err error
		answer, err = q.CreateAnswer(ctx, db.CreateAnswerParams{
			QuestionID: questionID,
			QuizID:     sql.NullInt64{Int64: quiz.ID, Valid: true},
			Response:   string(rawResp),
			Score:      score,
			Feedback:   feedback,
		})
		if err != nil {
			return err
		}
		if err := s.progress.Review(ctx, q, questionID, score, hinted, time.Now()); err != nil {
			return err
		}
		if err := q.ClearQuizDraft(ctx, db.ClearQuizDraftParams{QuizID: quizID, QuestionID: questionID}); err != nil {
			return err
		}
		if len(answers)+1 == len(questions) {
			completed = true
			if err := q.CompleteQuiz(ctx, quiz.ID); err != nil {
				return err
			}
			return s.settle(ctx, q, quiz, questions, append(answers, answer))
		}
		return nil
	})
	if err != nil {
		return Result{}, false, err
	}
	res, err := result(*question, answer)
	return res, completed, err
}

func (s *Service) settle(ctx context.Context, q *db.Queries, quiz db.Quiz, questions []db.Question, answers []db.Answer) error {
	now := time.Now()
	if quiz.Checkpoint && *averageScore(answers) >= progress.CheckpointPass {
		if err := s.progress.MarkLearned(ctx, q, quiz.TopicSlug.String, now); err != nil {
			return err
		}
	}
	var slugs []string
	seen := map[string]bool{}
	for _, question := range questions {
		if !seen[question.TopicSlug] {
			seen[question.TopicSlug] = true
			slugs = append(slugs, question.TopicSlug)
		}
	}
	return s.progress.Settle(ctx, q, slugs, now)
}

type Filter struct {
	TopicSlug string
	Open      *bool
	Limit     int64
}

func (s *Service) List(ctx context.Context, f Filter) ([]Summary, error) {
	params := db.ListQuizzesParams{RowLimit: f.Limit}
	if f.TopicSlug != "" {
		params.TopicSlug = f.TopicSlug
	}
	if f.Open != nil {
		params.Open = *f.Open
	}
	quizzes, err := s.q.ListQuizzes(ctx, params)
	if err != nil {
		return nil, err
	}
	out := make([]Summary, 0, len(quizzes))
	for _, quiz := range quizzes {
		total, err := s.q.CountQuizItems(ctx, quiz.ID)
		if err != nil {
			return nil, err
		}
		answers, err := s.q.ListQuizAnswers(ctx, sql.NullInt64{Int64: quiz.ID, Valid: true})
		if err != nil {
			return nil, err
		}
		out = append(out, Summary{
			ID:          quiz.ID,
			Kind:        kindOf(quiz),
			TopicSlug:   ptrString(quiz.TopicSlug),
			DailyDate:   ptrString(quiz.DailyDate),
			CreatedAt:   quiz.CreatedAt,
			CompletedAt: ptrInt(quiz.CompletedAt),
			Total:       total,
			Answered:    len(answers),
			Score:       averageScore(answers),
		})
	}
	return out, nil
}

func (s *Service) SaveDraft(ctx context.Context, quizID, questionID int64, resp Response) error {
	raw, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	n, err := s.q.SaveQuizDraft(ctx, db.SaveQuizDraftParams{Draft: string(raw), QuizID: quizID, QuestionID: questionID})
	if err != nil {
		return err
	}
	if n == 0 {
		return httpx.Errorf(http.StatusNotFound, "question %d is not part of quiz %d", questionID, quizID)
	}
	return nil
}

func kindOf(q db.Quiz) string {
	if q.Checkpoint {
		return quizCheckpoint
	}
	return q.Kind
}

func toHint(h db.Hint) Hint {
	return Hint{ID: h.ID, Kind: h.Kind, Content: h.Content, CreatedAt: h.CreatedAt}
}

func result(q db.Question, a db.Answer) (Result, error) {
	res := Result{
		Score:           a.Score,
		Correct:         a.Score >= progress.PassScore,
		Feedback:        a.Feedback,
		Explanation:     q.Explanation,
		ReferenceAnswer: q.ReferenceAnswer,
	}
	if err := json.Unmarshal([]byte(a.Response), &res.Response); err != nil {
		return res, err
	}
	if err := json.Unmarshal([]byte(q.Correct), &res.CorrectOptions); err != nil {
		return res, err
	}
	return res, nil
}

func averageScore(answers []db.Answer) *float64 {
	if len(answers) == 0 {
		return nil
	}
	var total float64
	for _, a := range answers {
		total += a.Score
	}
	avg := total / float64(len(answers))
	return &avg
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func ptrString(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

func ptrInt(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	return &n.Int64
}
