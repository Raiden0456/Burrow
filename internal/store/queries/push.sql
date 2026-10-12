-- name: UpsertPushSubscription :exec
INSERT INTO push_subscriptions (endpoint, p256dh, auth) VALUES (?, ?, ?)
ON CONFLICT (endpoint) DO UPDATE SET p256dh = excluded.p256dh, auth = excluded.auth;

-- name: DeletePushSubscription :exec
DELETE FROM push_subscriptions WHERE endpoint = ?;

-- name: ListPushSubscriptions :many
SELECT * FROM push_subscriptions;

-- name: CountPushSubscriptions :one
SELECT COUNT(*) FROM push_subscriptions;
