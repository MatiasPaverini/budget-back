package account

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/mpaverini/budget-back/internal/db"
	"github.com/mpaverini/budget-back/internal/pgutil"
)

type Type string

const (
	TypeCash       Type = "cash"
	TypeBank       Type = "bank"
	TypeCreditCard Type = "credit_card"
	TypeCreditLine Type = "credit_line"
	TypeInvestment Type = "investment"
	TypeLoan       Type = "loan"
)

func validType(t Type) bool {
	switch t {
	case TypeCash, TypeBank, TypeCreditCard, TypeCreditLine, TypeInvestment, TypeLoan:
		return true
	default:
		return false
	}
}

type Account struct {
	ID                uuid.UUID        `json:"id"`
	Name              string           `json:"name"`
	Type              Type             `json:"type"`
	Currency          string           `json:"currency"`
	OpeningBalance    decimal.Decimal  `json:"opening_balance"`
	OpenedAt          time.Time        `json:"opened_at"`
	CreditLimit       *decimal.Decimal `json:"credit_limit,omitempty"`
	StatementCloseDay *int16           `json:"statement_close_day,omitempty"`
	DueDay            *int16           `json:"due_day,omitempty"`
	InterestRate      *decimal.Decimal `json:"interest_rate,omitempty"`
	TermMonths        *int16           `json:"term_months,omitempty"`
}

type CreateInput struct {
	Name              string
	Type              Type
	Currency          string
	OpeningBalance    decimal.Decimal
	OpenedAt          time.Time
	CreditLimit       *decimal.Decimal
	StatementCloseDay *int16
	DueDay            *int16
	InterestRate      *decimal.Decimal
	TermMonths        *int16
}

type UpdateInput struct {
	Name              string
	Currency          string
	CreditLimit       *decimal.Decimal
	StatementCloseDay *int16
	DueDay            *int16
	InterestRate      *decimal.Decimal
	TermMonths        *int16
}

type Service struct {
	q *db.Queries
}

func NewService(q *db.Queries) *Service {
	return &Service{q: q}
}

func (s *Service) Create(ctx context.Context, userID string, in CreateInput) (Account, error) {
	if in.Name == "" {
		return Account{}, fmt.Errorf("name is required")
	}
	if !validType(in.Type) {
		return Account{}, fmt.Errorf("invalid account type %q", in.Type)
	}
	currency := in.Currency
	if currency == "" {
		currency = "ARS"
	}
	openedAt := in.OpenedAt
	if openedAt.IsZero() {
		openedAt = time.Now()
	}

	row, err := s.q.CreateAccount(ctx, db.CreateAccountParams{
		UserID:            userID,
		Name:              in.Name,
		Type:              db.AccountType(in.Type),
		Currency:          currency,
		OpeningBalance:    in.OpeningBalance,
		OpenedAt:          pgutil.Date(openedAt),
		CreditLimit:       pgutil.NullDecimal(in.CreditLimit),
		StatementCloseDay: pgutil.Int2(in.StatementCloseDay),
		DueDay:            pgutil.Int2(in.DueDay),
		InterestRate:      pgutil.NullDecimal(in.InterestRate),
		TermMonths:        pgutil.Int2(in.TermMonths),
	})
	if err != nil {
		return Account{}, err
	}
	return fromDB(row), nil
}

func (s *Service) Get(ctx context.Context, userID string, id uuid.UUID) (Account, error) {
	row, err := s.q.GetAccount(ctx, db.GetAccountParams{ID: pgutil.UUID(id), UserID: userID})
	if err != nil {
		return Account{}, err
	}
	return fromDB(row), nil
}

func (s *Service) List(ctx context.Context, userID string) ([]Account, error) {
	rows, err := s.q.ListAccounts(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Account, 0, len(rows))
	for _, r := range rows {
		out = append(out, fromDB(r))
	}
	return out, nil
}

func (s *Service) Update(ctx context.Context, userID string, id uuid.UUID, in UpdateInput) (Account, error) {
	if in.Name == "" {
		return Account{}, fmt.Errorf("name is required")
	}
	currency := in.Currency
	if currency == "" {
		currency = "ARS"
	}
	row, err := s.q.UpdateAccount(ctx, db.UpdateAccountParams{
		ID:                pgutil.UUID(id),
		UserID:            userID,
		Name:              in.Name,
		Currency:          currency,
		CreditLimit:       pgutil.NullDecimal(in.CreditLimit),
		StatementCloseDay: pgutil.Int2(in.StatementCloseDay),
		DueDay:            pgutil.Int2(in.DueDay),
		InterestRate:      pgutil.NullDecimal(in.InterestRate),
		TermMonths:        pgutil.Int2(in.TermMonths),
	})
	if err != nil {
		return Account{}, err
	}
	return fromDB(row), nil
}

func (s *Service) Delete(ctx context.Context, userID string, id uuid.UUID) error {
	return s.q.DeleteAccount(ctx, db.DeleteAccountParams{ID: pgutil.UUID(id), UserID: userID})
}

func fromDB(a db.Account) Account {
	return Account{
		ID:                pgutil.ToUUID(a.ID),
		Name:              a.Name,
		Type:              Type(a.Type),
		Currency:          a.Currency,
		OpeningBalance:    a.OpeningBalance,
		OpenedAt:          pgutil.ToTime(a.OpenedAt),
		CreditLimit:       pgutil.FromNullDecimal(a.CreditLimit),
		StatementCloseDay: pgutil.FromInt2(a.StatementCloseDay),
		DueDay:            pgutil.FromInt2(a.DueDay),
		InterestRate:      pgutil.FromNullDecimal(a.InterestRate),
		TermMonths:        pgutil.FromInt2(a.TermMonths),
	}
}
