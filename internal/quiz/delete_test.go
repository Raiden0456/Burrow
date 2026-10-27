package quiz

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/raiden0456/gonteractive/internal/store"
	"github.com/raiden0456/gonteractive/internal/store/db"
)

func TestDeleteKeepsAnswers(t *testing.T) {
	ctx := t.Context()
	conn, err := store.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	q := db.New(conn)
	s := &Service{conn: conn, q: q}

	if _, err := conn.ExecContext(ctx, `INSERT INTO questions (id, topic_slug, kind, prompt, explanation, difficulty) VALUES (1, 't', 'open', 'p', 'e', 1)`); err != nil {
		t.Fatal(err)
	}
	quiz, err := q.CreateQuiz(ctx, db.CreateQuizParams{Kind: quizDaily, DailyDate: nullString("2026-10-20")})
	if err != nil {
		t.Fatal(err)
	}
	if err := q.AddQuizItem(ctx, db.AddQuizItemParams{QuizID: quiz.ID, Position: 0, QuestionID: 1}); err != nil {
		t.Fatal(err)
	}
	answer, err := q.CreateAnswer(ctx, db.CreateAnswerParams{QuestionID: 1, QuizID: sql.NullInt64{Int64: quiz.ID, Valid: true}, Response: "{}", Score: 1})
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Delete(ctx, quiz.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := q.GetQuiz(ctx, quiz.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("quiz still there: %v", err)
	}
	if n, _ := q.CountQuizItems(ctx, quiz.ID); n != 0 {
		t.Fatalf("items left: %d", n)
	}
	var quizID sql.NullInt64
	if err := conn.QueryRowContext(ctx, `SELECT quiz_id FROM answers WHERE id = ?`, answer.ID).Scan(&quizID); err != nil {
		t.Fatalf("answer gone: %v", err)
	}
	if quizID.Valid {
		t.Fatalf("answer still linked to quiz %d", quizID.Int64)
	}
	if err := s.Delete(ctx, quiz.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("second delete: %v", err)
	}
}
