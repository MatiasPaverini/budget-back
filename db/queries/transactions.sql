-- name: CreateTransaction :one
INSERT INTO transactions (
    user_id, account_id, occurred_at, amount, description, category, transfer_account_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: ListTransactionsByAccount :many
SELECT * FROM transactions
WHERE user_id = $1 AND account_id = $2
  AND (sqlc.narg('from_date')::date IS NULL OR occurred_at >= sqlc.narg('from_date'))
  AND (sqlc.narg('to_date')::date IS NULL OR occurred_at <= sqlc.narg('to_date'))
ORDER BY occurred_at DESC;

-- name: ListTransactions :many
SELECT * FROM transactions
WHERE user_id = $1
  AND (sqlc.narg('from_date')::date IS NULL OR occurred_at >= sqlc.narg('from_date'))
  AND (sqlc.narg('to_date')::date IS NULL OR occurred_at <= sqlc.narg('to_date'))
ORDER BY occurred_at DESC;

-- name: UpdateTransaction :one
UPDATE transactions SET
    occurred_at = $3,
    amount = $4,
    description = $5,
    category = $6
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = $1 AND user_id = $2;

-- name: NetWorth :many
SELECT
    a.id AS account_id,
    a.name AS name,
    a.type AS type,
    a.currency AS currency,
    (a.opening_balance + COALESCE(SUM(t.amount), 0))::numeric(18, 2) AS balance
FROM accounts a
LEFT JOIN transactions t ON t.account_id = a.id
WHERE a.user_id = $1
GROUP BY a.id, a.name, a.type, a.currency, a.opening_balance
ORDER BY a.name;
