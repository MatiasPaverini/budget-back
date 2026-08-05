package httpapi

import (
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	"github.com/mpaverini/budget-back/internal/calculator"
	"github.com/mpaverini/budget-back/internal/indicator"
)

type calculatorHandlers struct {
	indicators *indicator.Service
}

func (h calculatorHandlers) register(mux *http.ServeMux) {
	mux.HandleFunc("POST /calculators/rent-projection", h.rentProjection)
	mux.HandleFunc("POST /calculators/purchase", h.purchase)
}

type rentProjectionRequest struct {
	BaseAmount                decimal.Decimal `json:"base_amount"`
	BasePeriod                time.Time       `json:"base_period"`
	AdjustmentFrequencyMonths int             `json:"adjustment_frequency_months"`
	TargetPeriod              *time.Time      `json:"target_period"`
}

func (h calculatorHandlers) rentProjection(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireUserID(w, r); !ok {
		return
	}
	var req rentProjectionRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	history, err := h.indicators.History(r.Context(), indicator.IPCCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load IPC history")
		return
	}

	result, err := calculator.ProjectRent(history, calculator.RentProjectionInput{
		BaseAmount:                req.BaseAmount,
		BasePeriod:                req.BasePeriod,
		AdjustmentFrequencyMonths: req.AdjustmentFrequencyMonths,
		TargetPeriod:              req.TargetPeriod,
	})
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

type purchaseRequest struct {
	Price               decimal.Decimal  `json:"price"`
	Installments        int              `json:"installments"`
	MonthlyInflationPct *decimal.Decimal `json:"monthly_inflation_pct"`
}

func (h calculatorHandlers) purchase(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireUserID(w, r); !ok {
		return
	}
	var req purchaseRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rate := req.MonthlyInflationPct
	if rate == nil {
		history, err := h.indicators.History(r.Context(), indicator.IPCCode)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load IPC history")
			return
		}
		avg, err := calculator.RecentMonthlyInflation(history, 3)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "no monthly_inflation_pct given and not enough IPC history to estimate one")
			return
		}
		rate = &avg
	}

	result, err := calculator.ProjectInstallmentPurchase(calculator.InstallmentPurchaseInput{
		Price:               req.Price,
		Installments:        req.Installments,
		MonthlyInflationPct: *rate,
	})
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}
