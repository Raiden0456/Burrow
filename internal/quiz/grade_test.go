package quiz

import (
	"slices"
	"testing"

	"github.com/raiden0456/gonteractive/internal/store/db"
)

func TestGradeChoice(t *testing.T) {
	single := db.Question{Kind: string(KindSingle), Correct: "[2]"}
	multi := db.Question{Kind: string(KindMulti), Correct: "[0, 2, 3]"}

	tests := []struct {
		name     string
		q        db.Question
		selected []int
		want     float64
	}{
		{"single correct", single, []int{2}, 1},
		{"single wrong", single, []int{1}, 0},
		{"single with extra pick", single, []int{1, 2}, 0},
		{"multi exact", multi, []int{3, 0, 2}, 1},
		{"multi partial", multi, []int{0, 2}, 2.0 / 3},
		{"multi with wrong pick", multi, []int{0, 1}, 0},
		{"multi nothing", multi, nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gradeChoice(tt.q, tt.selected)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeneratedSetValidate(t *testing.T) {
	valid := generatedQuestion{Kind: KindSingle, Prompt: "p", Options: []string{"a", "b", "c"}, Correct: []int{1}, Explanation: "e", Difficulty: 1}
	open := generatedQuestion{Kind: KindOpen, Prompt: "p", Options: []string{}, Correct: []int{}, ReferenceAnswer: "r", Explanation: "e", Difficulty: 2}

	tests := []struct {
		name    string
		qs      []generatedQuestion
		want    int
		wantErr bool
	}{
		{"ok and truncated", []generatedQuestion{valid, open, valid}, 2, false},
		{"too few", []generatedQuestion{valid}, 2, true},
		{"index out of range", []generatedQuestion{{Kind: KindSingle, Prompt: "p", Options: []string{"a", "b", "c"}, Correct: []int{3}, Explanation: "e"}}, 1, true},
		{"single with two answers", []generatedQuestion{{Kind: KindSingle, Prompt: "p", Options: []string{"a", "b", "c"}, Correct: []int{0, 1}, Explanation: "e"}}, 1, true},
		{"multi with one answer", []generatedQuestion{{Kind: KindMulti, Prompt: "p", Options: []string{"a", "b", "c"}, Correct: []int{0}, Explanation: "e"}}, 1, true},
		{"open without reference", []generatedQuestion{{Kind: KindOpen, Prompt: "p", Explanation: "e"}}, 1, true},
		{"open with options", []generatedQuestion{{Kind: KindOpen, Prompt: "p", Options: []string{"a"}, ReferenceAnswer: "r", Explanation: "e"}}, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := generatedSet{Questions: tt.qs, want: tt.want}
			err := set.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && len(set.Questions) != tt.want {
				t.Fatalf("got %d questions, want %d", len(set.Questions), tt.want)
			}
		})
	}
}

func TestDailyPlan(t *testing.T) {
	tests := []struct {
		name             string
		fresh, weakShare int
		frontier, weak   string
		want             []topicShare
	}{
		{"frontier and weak", 6, 2, "f", "w", []topicShare{{"f", 4}, {"w", 2}}},
		{"no practiced topics", 6, 2, "f", "", []topicShare{{"f", 6}}},
		{"everything learned", 6, 2, "", "w", []topicShare{{"w", 6}}},
		{"frontier keeps a slot when space is tight", 2, 2, "f", "w", []topicShare{{"f", 1}, {"w", 1}}},
		{"no room", 0, 2, "f", "w", nil},
		{"nothing to pick", 6, 2, "", "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dailyPlan(tt.fresh, tt.weakShare, tt.frontier, tt.weak)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
