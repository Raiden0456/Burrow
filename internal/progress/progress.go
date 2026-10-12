package progress

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"

	"github.com/raiden0456/gonteractive/internal/curriculum"
	"github.com/raiden0456/gonteractive/internal/httpx"
	"github.com/raiden0456/gonteractive/internal/store/db"
)

const (
	PassScore      = 0.7
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
}

type Service struct {
	q          *db.Queries
	curriculum *curriculum.Curriculum
	fsrs       *fsrs.FSRS
}

func New(q *db.Queries, c *curriculum.Curriculum) *Service {
	return &Service{q: q, curriculum: c, fsrs: fsrs.NewFSRS(fsrs.DefaultParam())}
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
	default:
		return fsrs.Good
	}
}

func (s *Service) Review(ctx context.Context, q *db.Queries, questionID int64, score float64, now time.Time) error {
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

	next := s.fsrs.Next(card, now, Rating(score)).Card
	state, err := json.Marshal(next)
	if err != nil {
		return err
	}
	return q.UpsertCard(ctx, db.UpsertCardParams{QuestionID: questionID, Due: next.Due.Unix(), State: string(state)})
}

func (s *Service) TopicStats(ctx context.Context) (map[string]*TopicStat, error) {
	history, err := s.q.ListAnswerHistory(ctx)
	if err != nil {
		return nil, err
	}
	solved, err := s.q.CountSolvedDrillsByTopic(ctx)
	if err != nil {
		return nil, err
	}
	return topicStats(history, solved), nil
}

func (s *Service) Summary(ctx context.Context, now time.Time) (Summary, error) {
	history, err := s.q.ListAnswerHistory(ctx)
	if err != nil {
		return Summary{}, err
	}
	solved, err := s.q.CountSolvedDrillsByTopic(ctx)
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

	stats := topicStats(history, solved)
	for _, slug := range s.curriculum.Ordered() {
		if st, ok := stats[slug]; ok {
			sum.Topics = append(sum.Topics, *st)
			sum.DrillsSolved += st.DrillsSolved
		}
	}
	return sum, nil
}

func (s *Service) WeakestActiveTopics(ctx context.Context, limit int) ([]string, error) {
	stats, err := s.TopicStats(ctx)
	if err != nil {
		return nil, err
	}

	active := make([]*TopicStat, 0, len(stats))
	for _, st := range stats {
		if _, ok := s.curriculum.Topic(st.Slug); ok {
			active = append(active, st)
		}
	}
	sort.Slice(active, func(i, j int) bool { return active[i].Mastery < active[j].Mastery })

	var out []string
	for _, st := range active {
		if len(out) == limit {
			return out, nil
		}
		out = append(out, st.Slug)
	}
	for _, slug := range s.curriculum.Ordered() {
		if len(out) == limit {
			break
		}
		if _, seen := stats[slug]; !seen {
			out = append(out, slug)
		}
	}
	return out, nil
}

func topicStats(history []db.ListAnswerHistoryRow, solved []db.CountSolvedDrillsByTopicRow) map[string]*TopicStat {
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
		st.Mastery = mastery(scores, st.DrillsSolved)
	}
	return stats
}

func mastery(latestScores map[int64]float64, drillsSolved int) float64 {
	var quizPart float64
	if n := len(latestScores); n > 0 {
		var total float64
		for _, s := range latestScores {
			total += s
		}
		quizPart = total / float64(n) * math.Min(1, float64(n)/coverageTarget)
	}
	drillPart := math.Min(1, float64(drillsSolved)/drillTarget)
	return math.Round((0.8*quizPart+0.2*drillPart)*100) / 100
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
