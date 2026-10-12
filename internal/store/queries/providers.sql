-- name: ListProviders :many
SELECT * FROM providers ORDER BY id;

-- name: GetProvider :one
SELECT * FROM providers WHERE id = ?;

-- name: CreateProvider :one
INSERT INTO providers (name, kind, config) VALUES (?, ?, ?) RETURNING *;

-- name: UpdateProvider :one
UPDATE providers SET name = ?, kind = ?, config = ? WHERE id = ? RETURNING *;

-- name: DeleteProvider :exec
DELETE FROM providers WHERE id = ?;
