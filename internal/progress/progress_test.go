package progress

import (
	"testing"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

func TestRating(t *testing.T) {
	tests := map[float64]fsrs.Rating{0: fsrs.Again, 0.49: fsrs.Again, 0.5: fsrs.Hard, 0.79: fsrs.Hard, 0.8: fsrs.Good, 1: fsrs.Good}
	for score, want := range tests {
		if got := Rating(score); got != want {
			t.Errorf("Rating(%v) = %v, want %v", score, got, want)
		}
	}
}

func TestMastery(t *testing.T) {
	tests := []struct {
		name   string
		scores map[int64]float64
		drills int
		want   float64
	}{
		{"nothing", nil, 0, 0},
		{"few perfect answers are capped by coverage", map[int64]float64{1: 1, 2: 1}, 0, 0.2},
		{"full coverage", map[int64]float64{1: 1, 2: 1, 3: 1, 4: 1, 5: 0, 6: 0, 7: 1, 8: 1}, 0, 0.6},
		{"drills only", nil, 3, 0.2},
		{"everything", map[int64]float64{1: 1, 2: 1, 3: 1, 4: 1, 5: 1, 6: 1, 7: 1, 8: 1}, 5, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mastery(tt.scores, tt.drills); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStreak(t *testing.T) {
	now := time.Date(2026, 10, 12, 20, 0, 0, 0, time.UTC)
	day := func(offset int) string { return now.AddDate(0, 0, -offset).Format(dayLayout) }

	tests := []struct {
		name string
		days map[string]int
		want int
	}{
		{"no activity", map[string]int{}, 0},
		{"today only", map[string]int{day(0): 1}, 1},
		{"yesterday keeps the streak alive", map[string]int{day(1): 2, day(2): 1}, 2},
		{"gap breaks it", map[string]int{day(0): 1, day(1): 1, day(3): 1}, 2},
		{"two days ago is broken", map[string]int{day(2): 1}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := streak(tt.days, now); got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}
