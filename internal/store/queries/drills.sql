-- name: CreateDrill :one
INSERT INTO drills (topic_slug, title, description, starter_code, test_code, solution, difficulty)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetDrill :one
SELECT * FROM drills WHERE id = ?;

-- name: ListDrillsByTopic :many
SELECT * FROM drills WHERE topic_slug = ? ORDER BY id DESC;

-- name: ListDrillTitlesByTopic :many
SELECT title FROM drills WHERE topic_slug = ? ORDER BY id DESC LIMIT ?;

-- name: MarkDrillSolved :exec
UPDATE drills SET solved_at = unixepoch() WHERE id = ? AND solved_at IS NULL;

-- name: CreateDrillAttempt :one
INSERT INTO drill_attempts (drill_id, code, passed, output) VALUES (?, ?, ?, ?) RETURNING *;

-- name: GetLatestDrillAttempt :one
SELECT * FROM drill_attempts WHERE drill_id = ? ORDER BY id DESC LIMIT 1;
