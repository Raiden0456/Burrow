-- name: GetLesson :one
SELECT * FROM lessons WHERE topic_slug = ?;

-- name: UpsertLesson :exec
INSERT INTO lessons (topic_slug, content) VALUES (?, ?)
ON CONFLICT (topic_slug) DO UPDATE SET content = excluded.content, created_at = unixepoch();

-- name: ListLessons :many
SELECT topic_slug, created_at FROM lessons ORDER BY topic_slug;
