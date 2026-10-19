package progress

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"

	"github.com/raiden0456/gonteractive/internal/curriculum"
	"github.com/raiden0456/gonteractive/internal/httpx"
	"github.com/raiden0456/gonteractive/internal/store/db"
)

const (
	PassScore      = 0.7
	LearnedMastery = 0.7
	CheckpointPass = 0.8
	coverageTarget = 8.0
	drillTarget    = 3.0
	activityDays   = 28
	dayLayout      = "2006-01-02"
)

type TopicStat struct {
	Slug         string  `json:"slug"`
	Answered     int     `json:"answered"`
	Mastery      float64 `json:"mastery"`
	DrillsSolved int     `json:"drillsSolved"`
	LastActivity int64   `json:"lastActivity"`
	Learned      bool    `json:"learned"`
}

type DayActivity struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type Summary struct {
	Streak       int           `json:"streak"`
	ActiveToday  bool          `json:"activeToday"`
	TotalAnswers int           `json:"totalAnswers"`
	AnswersToday int           `json:"answersToday"`
	Accuracy7d   *float64      `json:"accuracy7d"`
	DueCount     int64         `json:"dueCount"`
	DrillsSolved int           `json:"drillsSolved"`
	Activity     []DayActivity `json:"activity"`
	Topics       []TopicStat   `json:"topics"`
	Frontier     *string       `json:"frontier"`
}

type Service struct {
	q          *db.Queries
	curriculum *curriculum.Curriculum
	fsrs       *fsrs.FSRS
}

func New(q *db.Queries, c *curriculum.Curriculum) *Service {
	params := fsrs.DefaultParam()
	// Short-term steps reschedule within minutes, which a once-a-day quiz turns into "again tomorrow".
	params.EnableShortTerm = false
	return &Service{q: q, curriculum: c, fsrs: fsrs.NewFSRS(params)}
}

func (s *Service) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/progress", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		sum, err := s.Summary(r.Context(), time.Now())
		if err != nil {
			return err
		}
		httpx.JSON(w, http.StatusOK, sum)
		return nil
	}))
}

func Rating(score float64) fsrs.Rating {
	switch {
	case score < 0.5:
		return fsrs.Again
	case score < 0.8:
		return fsrs.Hard
	case score < 1:
		return fsrs.Good
	default:
		return fsrs.Easy
	}
}

func (s *Service) Review(ctx context.Context, q *db.Queries, questionID int64, score float64, hinted bool, now time.Time) error {
	card := fsrs.NewCard()
	row, err := q.GetCard(ctx, questionID)
	switch {
	case err == nil:
		if err := json.Unmarshal([]byte(row.State), &card); err != nil {
			return err
		}
	case !errors.Is(err, sql.ErrNoRows):
		return err
	}

	rating := Rating(score)
	if hinted {
		rating = min(rating, fsrs.Hard)
	}
	next := s.fsrs.Next(card, now, rating).Card
	state, err := json.Marshal(next)
	if err != nil {
		return err
	}
	return q.UpsertCard(ctx, db.UpsertCardParams{QuestionID: questionID, Due: next.Due.Unix(), State: string(state)})
}

func (s *Service) TopicStats(ctx context.Context, now time.Time) (map[string]*TopicStat, error) {
	return s.topicStats(ctx, s.q, now)
}

func (s *Service) topicStats(ctx context.Context, q *db.Queries, now time.Time) (map[string]*TopicStat, error) {
	history, err := q.ListAnswerHistory(ctx)
	if err != nil {
		return nil, err
	}
	return s.statsFrom(ctx, q, history, now)
}

func (s *Service) statsFrom(ctx context.Context, q *db.Queries, history []db.ListAnswerHistoryRow, now time.Time) (map[string]*TopicStat, error) {
	solved, err := q.CountSolvedDrillsByTopic(ctx)
	if err != nil {
		return nil, err
	}
	cards, err := q.ListCardStates(ctx)
	if err != nil {
		return nil, err
	}
	learned, err := q.ListLearnedTopics(ctx)
	if err != nil {
		return nil, err
	}

	retention := make(map[int64]float64, len(cards))
	for _, c := range cards {
		var card fsrs.Card
		if err := json.Unmarshal([]byte(c.State), &card); err != nil {
			return nil, err
		}
		retention[c.QuestionID] = s.fsrs.GetRetrievability(card, now)
	}

	stats := topicStats(history, retention, solved)
	for _, slug := range learned {
		if st, ok := stats[slug]; ok {
			st.Learned = true
		}
	}
	for _, st := range stats {
		st.Learned = st.Learned || st.Mastery >= LearnedMastery
	}
	return stats, nil
}

// Settle persists learned topics so that later forgetting lowers mastery without moving the frontier back.
func (s *Service) Settle(ctx context.Context, q *db.Queries, slugs []string, now time.Time) error {
	stats, err := s.topicStats(ctx, q, now)
	if err != nil {
		return err
	}
	for _, slug := range slugs {
		if st, ok := stats[slug]; ok && st.Mastery >= LearnedMastery {
			if err := s.MarkLearned(ctx, q, slug, now); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) MarkLearned(ctx context.Context, q *db.Queries, slug string, now time.Time) error {
	return q.MarkTopicLearned(ctx, db.MarkTopicLearnedParams{TopicSlug: slug, LearnedAt: now.Unix()})
}

func (s *Service) Summary(ctx context.Context, now time.Time) (Summary, error) {
	history, err := s.q.ListAnswerHistory(ctx)
	if err != nil {
		return Summary{}, err
	}
	stats, err := s.statsFrom(ctx, s.q, history, now)
	if err != nil {
		return Summary{}, err
	}
	solveTimes, err := s.q.ListDrillSolveTimes(ctx)
	if err != nil {
		return Summary{}, err
	}
	due, err := s.q.CountDueCards(ctx, now.Unix())
	if err != nil {
		return Summary{}, err
	}

	sum := Summary{TotalAnswers: len(history), DueCount: due, Topics: []TopicStat{}}

	perDay := map[string]int{}
	today := now.Format(dayLayout)
	weekAgo := now.AddDate(0, 0, -7).Unix()
	var weekScore float64
	var weekCount int
	for _, h := range history {
		day := time.Unix(h.CreatedAt, 0).In(now.Location()).Format(dayLayout)
		perDay[day]++
		if day == today {
			sum.AnswersToday++
		}
		if h.CreatedAt >= weekAgo {
			weekScore += h.Score
			weekCount++
		}
	}
	for _, t := range solveTimes {
		perDay[time.Unix(t.Int64, 0).In(now.Location()).Format(dayLayout)]++
	}
	if weekCount > 0 {
		acc := weekScore / float64(weekCount)
		sum.Accuracy7d = &acc
	}

	sum.ActiveToday = perDay[today] > 0
	sum.Streak = streak(perDay, now)
	for i := activityDays - 1; i >= 0; i-- {
		day := now.AddDate(0, 0, -i).Format(dayLayout)
		sum.Activity = append(sum.Activity, DayActivity{Date: day, Count: perDay[day]})
	}

	for _, slug := range s.curriculum.Ordered() {
		if st, ok := stats[slug]; ok {
			sum.Topics = append(sum.Topics, *st)
			sum.DrillsSolved += st.DrillsSolved
		}
	}
	if slug := s.frontier(stats); slug != "" {
		sum.Frontier = &slug
	}
	return sum, nil
}

func (s *Service) DailyTopics(ctx context.Context, now time.Time) (frontier, weakest string, err error) {
	stats, err := s.TopicStats(ctx, now)
	if err != nil {
		return "", "", err
	}
	frontier = s.frontier(stats)

	var weak *TopicStat
	for _, slug := range s.curriculum.Ordered() {
		st, ok := stats[slug]
		if !ok || slug == frontier || st.Answered == 0 {
			continue
		}
		if weak == nil || st.Mastery < weak.Mastery {
			weak = st
		}
	}
	if weak != nil {
		weakest = weak.Slug
	}
	return frontier, weakest, nil
}

func (s *Service) frontier(stats map[string]*TopicStat) string {
	for _, slug := range s.curriculum.Ordered() {
		if st, ok := stats[slug]; !ok || !st.Learned {
			return slug
		}
	}
	return ""
}

func topicStats(history []db.ListAnswerHistoryRow, retention map[int64]float64, solved []db.CountSolvedDrillsByTopicRow) map[string]*TopicStat {
	stats := map[string]*TopicStat{}
	get := func(slug string) *TopicStat {
		st, ok := stats[slug]
		if !ok {
			st = &TopicStat{Slug: slug}
			stats[slug] = st
		}
		return st
	}

	latest := map[string]map[int64]float64{}
	for _, h := range history {
		if latest[h.TopicSlug] == nil {
			latest[h.TopicSlug] = map[int64]float64{}
		}
		latest[h.TopicSlug][h.QuestionID] = h.Score
		st := get(h.TopicSlug)
		st.LastActivity = max(st.LastActivity, h.CreatedAt)
	}
	for _, d := range solved {
		get(d.TopicSlug).DrillsSolved = int(d.Solved)
	}

	for slug, st := range stats {
		scores := latest[slug]
		st.Answered = len(scores)
		recall := make(map[int64]float64, len(scores))
		for id, score := range scores {
			r, ok := retention[id]
			if !ok {
				r = 1
			}
			recall[id] = score * r
		}
		st.Mastery = mastery(recall, st.DrillsSolved)
	}
	return stats
}

func mastery(recall map[int64]float64, drillsSolved int) float64 {
	var quizPart float64
	if n := len(recall); n > 0 {
		var total float64
		for _, r := range recall {
			total += r
		}
		quizPart = total / float64(n) * math.Min(1, float64(n)/coverageTarget)
	}
	drillPart := math.Min(1, float64(drillsSolved)/drillTarget)
	return math.Round(math.Max(quizPart, 0.8*quizPart+0.2*drillPart)*100) / 100
}

func streak(perDay map[string]int, now time.Time) int {
	day := now
	if perDay[day.Format(dayLayout)] == 0 {
		day = day.AddDate(0, 0, -1)
	}
	n := 0
	for perDay[day.Format(dayLayout)] > 0 {
		n++
		day = day.AddDate(0, 0, -1)
	}
	return n
}
