package recurringcharge

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/mpaverini/budget-back/internal/db"
	"github.com/mpaverini/budget-back/internal/pgutil"
)

type RecurringCharge struct {
	ID                        uuid.UUID       `json:"id"`
	Name                      string          `json:"name"`
	BaseAmount                decimal.Decimal `json:"base_amount"`
	BasePeriod                time.Time       `json:"base_period"`
	AdjustmentFrequencyMonths int             `json:"adjustment_frequency_months"`
	IndexCode                 string          `json:"index_code"`
	NextReviewDate            time.Time       `json:"next_review_date"`
}

type CreateInput struct {
	Name                      string
	BaseAmount                decimal.Decimal
	BasePeriod                time.Time
	AdjustmentFrequencyMonths int
	IndexCode                 string
}

type UpdateInput struct {
	Name                      string
	BaseAmount                decimal.Decimal
	BasePeriod                time.Time
	AdjustmentFrequencyMonths int
	IndexCode                 string
}

type Service struct {
	q *db.Queries
}

func NewService(q *db.Queries) *Service {
	return &Service{q: q}
}

func (s *Service) Create(ctx context.Context, userID string, in CreateInput) (RecurringCharge, error) {
	if err := validate(in.Name, in.BasePeriod, in.AdjustmentFrequencyMonths, in.IndexCode); err != nil {
		return RecurringCharge{}, err
	}
	// next_review_date is always derived from base_period + frequency, never
	// taken from the caller — this keeps it in lockstep with the default
	// target period calculator.ProjectRent uses, so the stored "next review"
	// date and the projected amount shown for it never drift apart.
	nextReview := in.BasePeriod.AddDate(0, in.AdjustmentFrequencyMonths, 0)

	row, err := s.q.CreateRecurringCharge(ctx, db.CreateRecurringChargeParams{
		UserID:                    userID,
		Name:                      in.Name,
		BaseAmount:                in.BaseAmount,
		BasePeriod:                pgutil.Date(in.BasePeriod),
		AdjustmentFrequencyMonths: int16(in.AdjustmentFrequencyMonths),
		IndexCode:                 in.IndexCode,
		NextReviewDate:            pgutil.Date(nextReview),
	})
	if err != nil {
		return RecurringCharge{}, err
	}
	return fromDB(row), nil
}

func (s *Service) Get(ctx context.Context, userID string, id uuid.UUID) (RecurringCharge, error) {
	row, err := s.q.GetRecurringCharge(ctx, db.GetRecurringChargeParams{ID: pgutil.UUID(id), UserID: userID})
	if err != nil {
		return RecurringCharge{}, err
	}
	return fromDB(row), nil
}

func (s *Service) List(ctx context.Context, userID string) ([]RecurringCharge, error) {
	rows, err := s.q.ListRecurringCharges(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]RecurringCharge, 0, len(rows))
	for _, r := range rows {
		out = append(out, fromDB(r))
	}
	return out, nil
}

func (s *Service) Update(ctx context.Context, userID string, id uuid.UUID, in UpdateInput) (RecurringCharge, error) {
	if err := validate(in.Name, in.BasePeriod, in.AdjustmentFrequencyMonths, in.IndexCode); err != nil {
		return RecurringCharge{}, err
	}
	nextReview := in.BasePeriod.AddDate(0, in.AdjustmentFrequencyMonths, 0)

	row, err := s.q.UpdateRecurringCharge(ctx, db.UpdateRecurringChargeParams{
		ID:                        pgutil.UUID(id),
		UserID:                    userID,
		Name:                      in.Name,
		BaseAmount:                in.BaseAmount,
		BasePeriod:                pgutil.Date(in.BasePeriod),
		AdjustmentFrequencyMonths: int16(in.AdjustmentFrequencyMonths),
		IndexCode:                 in.IndexCode,
		NextReviewDate:            pgutil.Date(nextReview),
	})
	if err != nil {
		return RecurringCharge{}, err
	}
	return fromDB(row), nil
}

func (s *Service) Delete(ctx context.Context, userID string, id uuid.UUID) error {
	return s.q.DeleteRecurringCharge(ctx, db.DeleteRecurringChargeParams{ID: pgutil.UUID(id), UserID: userID})
}

func validate(name string, basePeriod time.Time, frequencyMonths int, indexCode string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if basePeriod.IsZero() {
		return fmt.Errorf("base_period is required")
	}
	if frequencyMonths <= 0 {
		return fmt.Errorf("adjustment_frequency_months must be positive")
	}
	if indexCode == "" {
		return fmt.Errorf("index_code is required")
	}
	return nil
}

func fromDB(r db.RecurringCharge) RecurringCharge {
	return RecurringCharge{
		ID:                        pgutil.ToUUID(r.ID),
		Name:                      r.Name,
		BaseAmount:                r.BaseAmount,
		BasePeriod:                pgutil.ToTime(r.BasePeriod),
		AdjustmentFrequencyMonths: int(r.AdjustmentFrequencyMonths),
		IndexCode:                 r.IndexCode,
		NextReviewDate:            pgutil.ToTime(r.NextReviewDate),
	}
}
