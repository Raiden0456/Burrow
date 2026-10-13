-- name: CreateJob :one
INSERT INTO jobs (kind, topic_slug, status) VALUES (?, ?, 'running') RETURNING *;

-- name: GetJob :one
SELECT * FROM jobs WHERE id = ?;

-- name: FindRunningJob :one
SELECT * FROM jobs WHERE kind = ? AND topic_slug = ? AND status = 'running' ORDER BY id DESC LIMIT 1;

-- name: FinishJob :exec
UPDATE jobs SET status = ?, result_id = ?, error = ?, finished_at = unixepoch() WHERE id = ?;

-- name: ListRecentJobs :many
SELECT * FROM jobs WHERE status = 'running' OR finished_at >= ? ORDER BY id DESC;

-- name: FailRunningJobs :exec
UPDATE jobs SET status = 'failed', error = ?, finished_at = unixepoch() WHERE status = 'running';
