-- name: ListSubscriptions :many
SELECT * FROM subscriptions ORDER BY name;

-- name: GetSubscription :one
SELECT * FROM subscriptions WHERE id = ?;

-- name: CreateSubscription :one
INSERT INTO subscriptions (name, youtube_id, type, thumbnail_url, position, active)
VALUES (?, ?, ?, ?, ?, ?)
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
INSERT INTO videos (subscription_id, youtube_id, title, thumbnail_url, duration, published_at, is_short)
VALUES (?, ?, ?, ?, ?, ?, ?)
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

-- name: ListAllSubscriptionsOrdered :many
SELECT s.*, COUNT(CASE WHEN v.watched = 0 THEN 1 END) as unwatched_count
FROM subscriptions s
LEFT JOIN videos v ON v.subscription_id = s.id
GROUP BY s.id
ORDER BY s.position, s.name;

-- name: ListActiveSubscriptions :many
SELECT s.*, COUNT(CASE WHEN v.watched = 0 THEN 1 END) as unwatched_count
FROM subscriptions s
LEFT JOIN videos v ON v.subscription_id = s.id
WHERE s.active = 1
GROUP BY s.id
ORDER BY s.position;

-- name: UpdateSubscriptionActive :exec
UPDATE subscriptions SET active = ? WHERE id = ?;

-- name: UpdateSubscriptionPosition :exec
UPDATE subscriptions SET position = ? WHERE id = ?;

-- name: GetMaxPosition :one
SELECT COALESCE(MAX(position), 0) as max_position FROM subscriptions;

-- name: FilterSubscriptions :many
SELECT s.*, COUNT(CASE WHEN v.watched = 0 THEN 1 END) as unwatched_count
FROM subscriptions s
LEFT JOIN videos v ON v.subscription_id = s.id
WHERE s.name LIKE '%' || ? || '%'
GROUP BY s.id
ORDER BY s.position, s.name
LIMIT 50;

-- name: ListSubscriptionsPaginated :many
SELECT s.*, COUNT(CASE WHEN v.watched = 0 THEN 1 END) as unwatched_count
FROM subscriptions s
LEFT JOIN videos v ON v.subscription_id = s.id
GROUP BY s.id
ORDER BY s.position, s.name
LIMIT ? OFFSET ?;

-- name: CountActiveSubscriptions :one
SELECT COUNT(*) FROM subscriptions WHERE active = 1;

-- name: ListUnwatchedVideosPaginated :many
SELECT * FROM videos
WHERE subscription_id = ? AND watched = 0
ORDER BY published_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateSubscriptionPageToken :exec
UPDATE subscriptions SET page_token = ? WHERE id = ?;

-- name: CountTotalVideos :one
SELECT COUNT(*) FROM videos WHERE subscription_id = ?;

-- name: UpdateSubscriptionHideShorts :exec
UPDATE subscriptions SET hide_shorts = ? WHERE id = ?;

-- name: ListUnwatchedVideosPaginatedFiltered :many
SELECT * FROM videos
WHERE subscription_id = ?
  AND watched = 0
  AND (? = 0 OR is_short = 0)
ORDER BY published_at DESC
LIMIT ? OFFSET ?;

-- name: CountUnwatchedBySubscriptionFiltered :one
SELECT COUNT(*) FROM videos
WHERE subscription_id = ?
  AND watched = 0
  AND (? = 0 OR is_short = 0);

-- name: UpdateVideoIsShort :exec
UPDATE videos SET is_short = ? WHERE id = ?;
