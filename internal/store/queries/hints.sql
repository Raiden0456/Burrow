-- name: CreateHint :one
INSERT INTO hints (kind, quiz_id, question_id, drill_id, content) VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: ListQuizHints :many
SELECT * FROM hints WHERE quiz_id = ? ORDER BY id;

-- name: ListQuestionHints :many
SELECT * FROM hints WHERE quiz_id = ? AND question_id = ? ORDER BY id;

-- name: ListDrillHints :many
SELECT * FROM hints WHERE drill_id = ? ORDER BY id;
