-- name: CreateRecurringCharge :one
INSERT INTO recurring_charges (
    user_id, name, base_amount, base_period, adjustment_frequency_months, index_code, next_review_date
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListRecurringCharges :many
SELECT * FROM recurring_charges WHERE user_id = $1 ORDER BY next_review_date;

-- name: GetRecurringCharge :one
SELECT * FROM recurring_charges WHERE id = $1 AND user_id = $2;

-- name: UpdateRecurringCharge :one
UPDATE recurring_charges SET
    name = $3,
    base_amount = $4,
    base_period = $5,
    adjustment_frequency_months = $6,
    index_code = $7,
    next_review_date = $8,
    updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteRecurringCharge :exec
DELETE FROM recurring_charges WHERE id = $1 AND user_id = $2;
