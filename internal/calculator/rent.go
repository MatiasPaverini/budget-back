package calculator

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/mpaverini/budget-back/internal/indicator"
)

type RentProjectionInput struct {
	BaseAmount                decimal.Decimal
	BasePeriod                time.Time
	AdjustmentFrequencyMonths int
	// TargetPeriod defaults to BasePeriod + AdjustmentFrequencyMonths when nil.
	TargetPeriod *time.Time
}

type RentProjectionResult struct {
	BasePeriod time.Time `json:"base_period"`
	// BaseIndexPeriod is the month BaseIndexValue actually came from — it
	// can differ from BasePeriod (see BaseEstimated).
	BaseIndexPeriod time.Time       `json:"base_index_period"`
	BaseIndexValue  decimal.Decimal `json:"base_index_value"`
	// BaseEstimated is true when there's no published IPC for BasePeriod yet
	// (e.g. it's the current month) and the latest known value before it
	// was used instead.
	BaseEstimated bool      `json:"base_estimated"`
	TargetPeriod  time.Time `json:"target_period"`
	// TargetIndexPeriod is the month TargetIndexValue actually came from —
	// see TargetEstimated.
	TargetIndexPeriod time.Time       `json:"target_index_period"`
	TargetIndexValue  decimal.Decimal `json:"target_index_value"`
	// TargetEstimated is true when TargetPeriod's IPC hasn't been published
	// yet and the latest known value was used as a stand-in. If both
	// BaseEstimated and TargetEstimated fall back to the *same* month, the
	// projection correctly shows 0% change — that means there's no new data
	// since BaseIndexPeriod, not that inflation is actually zero.
	TargetEstimated        bool            `json:"target_estimated"`
	ProjectedAmount        decimal.Decimal `json:"projected_amount"`
	CumulativeInflationPct decimal.Decimal `json:"cumulative_inflation_pct"`
}

// ProjectRent estimates what a periodically-adjusted charge (rent, a
// therapist's fee, anything bumped by IPC) should cost at TargetPeriod,
// given known IPC history, using the standard index-ratio escalation:
// newAmount = oldAmount × (index at target ÷ index at base) — the same
// formula ICL/UVA-style contract adjustments use. It's a pure function:
// history is whatever the caller already loaded from the indicator service.
func ProjectRent(history []indicator.Point, in RentProjectionInput) (RentProjectionResult, error) {
	if len(history) == 0 {
		return RentProjectionResult{}, fmt.Errorf("no IPC history available")
	}

	target := in.BasePeriod.AddDate(0, in.AdjustmentFrequencyMonths, 0)
	if in.TargetPeriod != nil {
		target = *in.TargetPeriod
	}

	baseAsOf, baseValue, baseExact, err := valueAtOrBefore(history, in.BasePeriod)
	if err != nil {
		return RentProjectionResult{}, fmt.Errorf("no IPC data at or before base period %s", in.BasePeriod.Format("2006-01"))
	}

	targetAsOf, targetValue, targetExact, err := valueAtOrBefore(history, target)
	if err != nil {
		return RentProjectionResult{}, fmt.Errorf("no IPC data available at or before target period %s", target.Format("2006-01"))
	}

	ratio := targetValue.Div(baseValue)
	projected := in.BaseAmount.Mul(ratio).Round(2)
	cumulativePct := ratio.Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100)).Round(2)

	return RentProjectionResult{
		BasePeriod:             in.BasePeriod,
		BaseIndexPeriod:        baseAsOf,
		BaseIndexValue:         baseValue,
		BaseEstimated:          !baseExact,
		TargetPeriod:           target,
		TargetIndexPeriod:      targetAsOf,
		TargetIndexValue:       targetValue,
		TargetEstimated:        !targetExact,
		ProjectedAmount:        projected,
		CumulativeInflationPct: cumulativePct,
	}, nil
}

// valueAtOrBefore returns the latest known index value at or before `at`,
// plus the period it actually came from, assuming history is sorted
// ascending by period (as ListIndicatorHistory returns it). exact reports
// whether a data point exists for that exact month, versus falling back to
// the latest earlier one.
func valueAtOrBefore(history []indicator.Point, at time.Time) (asOf time.Time, value decimal.Decimal, exact bool, err error) {
	found := false
	for _, p := range history {
		if p.Period.After(at) {
			break
		}
		asOf = p.Period
		value = p.Value
		found = true
		exact = sameMonth(p.Period, at)
	}
	if !found {
		return time.Time{}, decimal.Decimal{}, false, fmt.Errorf("no data at or before %s", at.Format("2006-01"))
	}
	return asOf, value, exact, nil
}

func sameMonth(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month()
}
