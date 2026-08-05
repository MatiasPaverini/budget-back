CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE account_type AS ENUM ('cash', 'bank', 'credit_card', 'credit_line', 'investment', 'loan');

CREATE TABLE accounts (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              TEXT NOT NULL,
    name                 TEXT NOT NULL,
    type                 account_type NOT NULL,
    currency             TEXT NOT NULL DEFAULT 'ARS',
    opening_balance      NUMERIC(18, 2) NOT NULL DEFAULT 0,
    opened_at            DATE NOT NULL DEFAULT CURRENT_DATE,
    credit_limit         NUMERIC(18, 2),
    statement_close_day  SMALLINT,
    due_day              SMALLINT,
    interest_rate        NUMERIC(6, 4),
    term_months          SMALLINT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_accounts_user_id ON accounts (user_id);

CREATE TABLE transactions (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              TEXT NOT NULL,
    account_id           UUID NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    occurred_at          DATE NOT NULL,
    amount               NUMERIC(18, 2) NOT NULL,
    description          TEXT NOT NULL DEFAULT '',
    category             TEXT,
    transfer_account_id  UUID REFERENCES accounts (id) ON DELETE SET NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_transactions_user_id ON transactions (user_id);
CREATE INDEX idx_transactions_account_id ON transactions (account_id);
CREATE INDEX idx_transactions_occurred_at ON transactions (occurred_at);

CREATE TABLE indicators (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        TEXT NOT NULL,
    period      DATE NOT NULL,
    value       NUMERIC(10, 4) NOT NULL,
    source      TEXT NOT NULL DEFAULT 'manual',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (code, period)
);

CREATE TABLE recurring_charges (
    id                            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                       TEXT NOT NULL,
    name                          TEXT NOT NULL,
    base_amount                   NUMERIC(18, 2) NOT NULL,
    base_period                   DATE NOT NULL,
    adjustment_frequency_months   SMALLINT NOT NULL,
    index_code                    TEXT NOT NULL,
    next_review_date              DATE NOT NULL,
    created_at                    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_recurring_charges_user_id ON recurring_charges (user_id);
