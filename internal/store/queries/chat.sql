-- name: CreateThread :one
INSERT INTO chat_threads (title, context) VALUES (?, ?) RETURNING *;

-- name: ListThreads :many
SELECT * FROM chat_threads ORDER BY updated_at DESC, id DESC LIMIT ?;

-- name: GetThread :one
SELECT * FROM chat_threads WHERE id = ?;

-- name: TouchThread :exec
UPDATE chat_threads SET updated_at = unixepoch() WHERE id = ?;

-- name: DeleteThread :exec
DELETE FROM chat_threads WHERE id = ?;

-- name: CreateMessage :one
INSERT INTO chat_messages (thread_id, role, content) VALUES (?, ?, ?) RETURNING *;

-- name: ListMessages :many
SELECT * FROM chat_messages WHERE thread_id = ? ORDER BY id;

-- name: UpdateThreadTitle :one
UPDATE chat_threads SET title = ? WHERE id = ? RETURNING *;
