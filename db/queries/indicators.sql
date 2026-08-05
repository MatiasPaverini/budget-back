-- name: UpsertIndicator :one
INSERT INTO indicators (code, period, value, source)
VALUES ($1, $2, $3, $4)
ON CONFLICT (code, period) DO UPDATE SET value = EXCLUDED.value, source = EXCLUDED.source
RETURNING *;

-- name: ListIndicatorHistory :many
SELECT * FROM indicators WHERE code = $1 ORDER BY period ASC;

-- name: LatestIndicator :one
SELECT * FROM indicators WHERE code = $1 ORDER BY period DESC LIMIT 1;
