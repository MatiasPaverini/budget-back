-- name: CreateAccount :one
INSERT INTO accounts (
    user_id, name, type, currency, opening_balance, opened_at,
    credit_limit, statement_close_day, due_day, interest_rate, term_months
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts WHERE id = $1 AND user_id = $2;

-- name: ListAccounts :many
SELECT * FROM accounts WHERE user_id = $1 ORDER BY created_at;

-- name: UpdateAccount :one
UPDATE accounts SET
    name = $3,
    currency = $4,
    credit_limit = $5,
    statement_close_day = $6,
    due_day = $7,
    interest_rate = $8,
    term_months = $9,
    updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts WHERE id = $1 AND user_id = $2;
