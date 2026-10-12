-- name: CreateQuiz :one
INSERT INTO quizzes (kind, topic_slug, daily_date) VALUES (?, ?, ?) RETURNING *;

-- name: AddQuizItem :exec
INSERT INTO quiz_items (quiz_id, position, question_id) VALUES (?, ?, ?);

-- name: GetQuiz :one
SELECT * FROM quizzes WHERE id = ?;

-- name: GetDailyQuiz :one
SELECT * FROM quizzes WHERE daily_date = ?;

-- name: ListRecentQuizzes :many
SELECT * FROM quizzes ORDER BY id DESC LIMIT ?;

-- name: ListQuizQuestions :many
SELECT q.*
FROM quiz_items i
JOIN questions q ON q.id = i.question_id
WHERE i.quiz_id = ?
ORDER BY i.position;

-- name: CountQuizItems :one
SELECT COUNT(*) FROM quiz_items WHERE quiz_id = ?;

-- name: ListQuizAnswers :many
SELECT * FROM answers WHERE quiz_id = ? ORDER BY id;

-- name: CreateAnswer :one
INSERT INTO answers (question_id, quiz_id, response, score, feedback)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: CompleteQuiz :exec
UPDATE quizzes SET completed_at = unixepoch() WHERE id = ? AND completed_at IS NULL;
