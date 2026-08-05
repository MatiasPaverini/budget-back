package calculator

import (
	"fmt"
	"math"

	"github.com/shopspring/decimal"

	"github.com/mpaverini/budget-back/internal/indicator"
)

type InstallmentPurchaseInput struct {
	Price               decimal.Decimal
	Installments        int
	MonthlyInflationPct decimal.Decimal
}

type InstallmentPurchaseResult struct {
	InstallmentAmount   decimal.Decimal `json:"installment_amount"`
	NominalTotal        decimal.Decimal `json:"nominal_total"`
	RealTotalTodayValue decimal.Decimal `json:"real_total_today_value"`
	// EffectiveDiscountPct is how much cheaper the purchase is in today's
	// money versus paying cash up front, since fixed peso installments lose
	// value to inflation the further out they're paid. Common in Argentina,
	// where "cuotas sin interés" are usually a real discount when inflation
	// is high.
	EffectiveDiscountPct decimal.Decimal `json:"effective_discount_pct"`
}

// ProjectInstallmentPurchase compares paying `Price` in N fixed nominal
// installments against paying cash today, discounting each future
// installment back to today's money at MonthlyInflationPct.
func ProjectInstallmentPurchase(in InstallmentPurchaseInput) (InstallmentPurchaseResult, error) {
	if in.Installments <= 0 {
		return InstallmentPurchaseResult{}, fmt.Errorf("installments must be greater than zero")
	}

	n := decimal.NewFromInt(int64(in.Installments))
	installmentAmount := in.Price.Div(n).Round(2)
	nominalTotal := installmentAmount.Mul(n)

	one := decimal.NewFromInt(1)
	monthlyRate := in.MonthlyInflationPct.Div(decimal.NewFromInt(100))

	realTotal := decimal.Zero
	discountFactor := one
	for i := 0; i < in.Installments; i++ {
		realTotal = realTotal.Add(installmentAmount.Div(discountFactor))
		discountFactor = discountFactor.Mul(one.Add(monthlyRate))
	}
	realTotal = realTotal.Round(2)

	discountPct := one.Sub(realTotal.Div(in.Price)).Mul(decimal.NewFromInt(100)).Round(2)

	return InstallmentPurchaseResult{
		InstallmentAmount:    installmentAmount,
		NominalTotal:         nominalTotal,
		RealTotalTodayValue:  realTotal,
		EffectiveDiscountPct: discountPct,
	}, nil
}

// RecentMonthlyInflation returns the geometric-mean month-over-month IPC
// change (as a percentage) over the last `months` intervals in history.
// Used as the purchase calculator's default rate when the caller doesn't
// supply one explicitly. This is a forward-looking assumption, not a
// stored monetary value, so float64 precision is fine here.
func RecentMonthlyInflation(history []indicator.Point, months int) (decimal.Decimal, error) {
	if len(history) < 2 {
		return decimal.Decimal{}, fmt.Errorf("not enough IPC history to compute an average")
	}
	n := months
	if n > len(history)-1 {
		n = len(history) - 1
	}
	if n < 1 {
		return decimal.Decimal{}, fmt.Errorf("not enough IPC history to compute an average")
	}

	start := history[len(history)-1-n]
	end := history[len(history)-1]

	ratio, _ := end.Value.Div(start.Value).Float64()
	monthlyRatio := math.Pow(ratio, 1/float64(n))
	return decimal.NewFromFloat((monthlyRatio - 1) * 100), nil
}
