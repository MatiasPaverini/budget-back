package indicator

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"github.com/mpaverini/budget-back/internal/db"
	"github.com/mpaverini/budget-back/internal/pgutil"
)

// IPCCode is the only index code this app tracks today (Argentina's
// Consumer Price Index). The table shape supports others later.
const IPCCode = "ipc"

type Point struct {
	Code   string          `json:"code"`
	Period time.Time       `json:"period"`
	Value  decimal.Decimal `json:"value"`
	Source string          `json:"source"`
}

type Service struct {
	q *db.Queries
}

func NewService(q *db.Queries) *Service {
	return &Service{q: q}
}

// RecordManual stores a user-entered value for a period, e.g. a therapist's
// price bump that has nothing to do with the published IPC series.
func (s *Service) RecordManual(ctx context.Context, code string, period time.Time, value decimal.Decimal) (Point, error) {
	return s.upsert(ctx, code, period, value, "manual")
}

// upsert is keyed on (code, period), so re-recording the same month is a
// no-op replace rather than a duplicate row.
func (s *Service) upsert(ctx context.Context, code string, period time.Time, value decimal.Decimal, source string) (Point, error) {
	row, err := s.q.UpsertIndicator(ctx, db.UpsertIndicatorParams{
		Code:   code,
		Period: pgutil.Date(firstOfMonth(period)),
		Value:  value,
		Source: source,
	})
	if err != nil {
		return Point{}, err
	}
	return fromDB(row), nil
}

func (s *Service) History(ctx context.Context, code string) ([]Point, error) {
	rows, err := s.q.ListIndicatorHistory(ctx, code)
	if err != nil {
		return nil, err
	}
	out := make([]Point, 0, len(rows))
	for _, r := range rows {
		out = append(out, fromDB(r))
	}
	return out, nil
}

func (s *Service) Latest(ctx context.Context, code string) (Point, error) {
	row, err := s.q.LatestIndicator(ctx, code)
	if err != nil {
		return Point{}, err
	}
	return fromDB(row), nil
}

func firstOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func fromDB(i db.Indicator) Point {
	return Point{
		Code:   i.Code,
		Period: pgutil.ToTime(i.Period),
		Value:  i.Value,
		Source: i.Source,
	}
}
