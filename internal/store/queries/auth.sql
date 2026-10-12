-- name: CreateSession :exec
INSERT INTO sessions (token_hash, expires_at) VALUES (?, ?);

-- name: CountActiveSessions :one
SELECT COUNT(*) FROM sessions WHERE token_hash = ? AND expires_at > ?;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token_hash = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at <= ?;
