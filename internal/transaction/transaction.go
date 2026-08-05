package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/mpaverini/budget-back/internal/db"
	"github.com/mpaverini/budget-back/internal/pgutil"
)

type Transaction struct {
	ID                uuid.UUID       `json:"id"`
	AccountID         uuid.UUID       `json:"account_id"`
	OccurredAt        time.Time       `json:"occurred_at"`
	Amount            decimal.Decimal `json:"amount"`
	Description       string          `json:"description"`
	Category          *string         `json:"category,omitempty"`
	TransferAccountID *uuid.UUID      `json:"transfer_account_id,omitempty"`
}

type CreateInput struct {
	AccountID         uuid.UUID
	OccurredAt        time.Time
	Amount            decimal.Decimal
	Description       string
	Category          *string
	TransferAccountID *uuid.UUID
}

// UpdateInput deliberately excludes AccountID and TransferAccountID — moving
// a transaction to a different account isn't supported (delete and
// recreate instead), matching how account.UpdateInput also excludes its
// own foundational fields (Type, OpeningBalance).
type UpdateInput struct {
	OccurredAt  time.Time
	Amount      decimal.Decimal
	Description string
	Category    *string
}

type ListFilter struct {
	AccountID *uuid.UUID
	From      *time.Time
	To        *time.Time
}

type Balance struct {
	AccountID uuid.UUID       `json:"account_id"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Currency  string          `json:"currency"`
	Balance   decimal.Decimal `json:"balance"`
}

type Service struct {
	q *db.Queries
}

func NewService(q *db.Queries) *Service {
	return &Service{q: q}
}

func (s *Service) Create(ctx context.Context, userID string, in CreateInput) (Transaction, error) {
	if in.AccountID == uuid.Nil {
		return Transaction{}, fmt.Errorf("account_id is required")
	}
	if in.OccurredAt.IsZero() {
		in.OccurredAt = time.Now()
	}

	row, err := s.q.CreateTransaction(ctx, db.CreateTransactionParams{
		UserID:            userID,
		AccountID:         pgutil.UUID(in.AccountID),
		OccurredAt:        pgutil.Date(in.OccurredAt),
		Amount:            in.Amount,
		Description:       in.Description,
		Category:          pgutil.Text(in.Category),
		TransferAccountID: pgutil.NullUUID(in.TransferAccountID),
	})
	if err != nil {
		return Transaction{}, err
	}
	return fromDB(row), nil
}

func (s *Service) List(ctx context.Context, userID string, f ListFilter) ([]Transaction, error) {
	fromDate, toDate := dateFilter(f.From), dateFilter(f.To)

	var rows []db.Transaction
	var err error
	if f.AccountID != nil {
		rows, err = s.q.ListTransactionsByAccount(ctx, db.ListTransactionsByAccountParams{
			UserID:    userID,
			AccountID: pgutil.UUID(*f.AccountID),
			FromDate:  fromDate,
			ToDate:    toDate,
		})
	} else {
		rows, err = s.q.ListTransactions(ctx, db.ListTransactionsParams{
			UserID:   userID,
			FromDate: fromDate,
			ToDate:   toDate,
		})
	}
	if err != nil {
		return nil, err
	}

	out := make([]Transaction, 0, len(rows))
	for _, r := range rows {
		out = append(out, fromDB(r))
	}
	return out, nil
}

func (s *Service) Update(ctx context.Context, userID string, id uuid.UUID, in UpdateInput) (Transaction, error) {
	if in.OccurredAt.IsZero() {
		in.OccurredAt = time.Now()
	}
	row, err := s.q.UpdateTransaction(ctx, db.UpdateTransactionParams{
		ID:          pgutil.UUID(id),
		UserID:      userID,
		OccurredAt:  pgutil.Date(in.OccurredAt),
		Amount:      in.Amount,
		Description: in.Description,
		Category:    pgutil.Text(in.Category),
	})
	if err != nil {
		return Transaction{}, err
	}
	return fromDB(row), nil
}

func (s *Service) Delete(ctx context.Context, userID string, id uuid.UUID) error {
	return s.q.DeleteTransaction(ctx, db.DeleteTransactionParams{ID: pgutil.UUID(id), UserID: userID})
}

func (s *Service) NetWorth(ctx context.Context, userID string) ([]Balance, error) {
	rows, err := s.q.NetWorth(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Balance, 0, len(rows))
	for _, r := range rows {
		out = append(out, Balance{
			AccountID: pgutil.ToUUID(r.AccountID),
			Name:      r.Name,
			Type:      string(r.Type),
			Currency:  r.Currency,
			Balance:   r.Balance,
		})
	}
	return out, nil
}

func dateFilter(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{}
	}
	return pgutil.Date(*t)
}

func fromDB(t db.Transaction) Transaction {
	return Transaction{
		ID:                pgutil.ToUUID(t.ID),
		AccountID:         pgutil.ToUUID(t.AccountID),
		OccurredAt:        pgutil.ToTime(t.OccurredAt),
		Amount:            t.Amount,
		Description:       t.Description,
		Category:          pgutil.FromText(t.Category),
		TransferAccountID: pgutil.ToNullUUID(t.TransferAccountID),
	}
}
