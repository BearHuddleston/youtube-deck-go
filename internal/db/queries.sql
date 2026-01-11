-- name: ListSubscriptions :many
SELECT * FROM subscriptions ORDER BY name;

-- name: GetSubscription :one
SELECT * FROM subscriptions WHERE id = ?;

-- name: CreateSubscription :one
INSERT INTO subscriptions (name, youtube_id, type, thumbnail_url)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: DeleteSubscription :exec
DELETE FROM subscriptions WHERE id = ?;

-- name: UpdateSubscriptionChecked :exec
UPDATE subscriptions SET last_checked = CURRENT_TIMESTAMP WHERE id = ?;

-- name: ListVideos :many
SELECT * FROM videos WHERE subscription_id = ? ORDER BY published_at DESC;

-- name: ListUnwatchedVideos :many
SELECT * FROM videos WHERE subscription_id = ? AND watched = 0 ORDER BY published_at DESC;

-- name: GetVideo :one
SELECT * FROM videos WHERE id = ?;

-- name: CreateVideo :one
INSERT INTO videos (subscription_id, youtube_id, title, thumbnail_url, duration, published_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ToggleWatched :one
UPDATE videos SET watched = NOT watched WHERE id = ?
RETURNING *;

-- name: MarkWatched :exec
UPDATE videos SET watched = 1 WHERE id = ?;

-- name: MarkUnwatched :exec
UPDATE videos SET watched = 0 WHERE id = ?;

-- name: CountUnwatchedBySubscription :one
SELECT COUNT(*) FROM videos WHERE subscription_id = ? AND watched = 0;

-- name: VideoExistsByYoutubeID :one
SELECT EXISTS(SELECT 1 FROM videos WHERE youtube_id = ?);

-- name: SubscriptionsWithUnwatchedCount :many
SELECT s.*, COUNT(CASE WHEN v.watched = 0 THEN 1 END) as unwatched_count
FROM subscriptions s
LEFT JOIN videos v ON v.subscription_id = s.id
GROUP BY s.id
ORDER BY s.name;

-- name: ListSubscriptionsWithUnwatchedCount :many
SELECT s.*, COUNT(CASE WHEN v.watched = 0 THEN 1 END) as unwatched_count
FROM subscriptions s
LEFT JOIN videos v ON v.subscription_id = s.id
GROUP BY s.id
ORDER BY s.name;
