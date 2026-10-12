-- name: CreateQuestion :one
INSERT INTO questions (topic_slug, kind, prompt, options, correct, reference_answer, explanation, difficulty)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetQuestion :one
SELECT * FROM questions WHERE id = ?;

-- name: ListRecentPromptsByTopic :many
SELECT prompt FROM questions WHERE topic_slug = ? ORDER BY id DESC LIMIT ?;

-- name: ListRecentMistakesByTopic :many
SELECT q.prompt, a.response, a.feedback
FROM answers a
JOIN questions q ON q.id = a.question_id
WHERE q.topic_slug = ? AND a.score < 0.7
ORDER BY a.id DESC
LIMIT ?;

-- name: GetCard :one
SELECT * FROM cards WHERE question_id = ?;

-- name: UpsertCard :exec
INSERT INTO cards (question_id, due, state) VALUES (?, ?, ?)
ON CONFLICT (question_id) DO UPDATE SET due = excluded.due, state = excluded.state;

-- name: ListDueQuestionIDs :many
SELECT question_id FROM cards WHERE due <= ? ORDER BY due LIMIT ?;

-- name: CountDueCards :one
SELECT COUNT(*) FROM cards WHERE due <= ?;
