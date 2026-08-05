package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/mpaverini/budget-back/internal/calculator"
	"github.com/mpaverini/budget-back/internal/indicator"
	"github.com/mpaverini/budget-back/internal/recurringcharge"
)

type recurringChargeHandlers struct {
	svc        *recurringcharge.Service
	indicators *indicator.Service
}

func (h recurringChargeHandlers) register(mux *http.ServeMux) {
	mux.HandleFunc("GET /recurring-charges", h.list)
	mux.HandleFunc("POST /recurring-charges", h.create)
	mux.HandleFunc("GET /recurring-charges/{id}", h.get)
	mux.HandleFunc("PATCH /recurring-charges/{id}", h.update)
	mux.HandleFunc("DELETE /recurring-charges/{id}", h.delete)
}

type recurringChargeRequest struct {
	Name                      string          `json:"name"`
	BaseAmount                decimal.Decimal `json:"base_amount"`
	BasePeriod                time.Time       `json:"base_period"`
	AdjustmentFrequencyMonths int             `json:"adjustment_frequency_months"`
	IndexCode                 string          `json:"index_code"`
}

// recurringChargeResponse embeds the stored charge and adds a preview of
// what it'll cost at NextReviewDate, computed the same way the standalone
// rent calculator does — the whole point of tracking a recurring charge here
// is seeing that number before the bill actually changes. Projection is
// omitted (with ProjectionError explaining why) rather than failing the
// whole request when index history isn't available yet for IndexCode.
type recurringChargeResponse struct {
	recurringcharge.RecurringCharge
	Projection      *calculator.RentProjectionResult `json:"projection,omitempty"`
	ProjectionError string                           `json:"projection_error,omitempty"`
}

func (h recurringChargeHandlers) withProjection(ctx context.Context, charge recurringcharge.RecurringCharge) recurringChargeResponse {
	resp := recurringChargeResponse{RecurringCharge: charge}

	history, err := h.indicators.History(ctx, charge.IndexCode)
	if err != nil {
		resp.ProjectionError = "failed to load index history"
		return resp
	}
	result, err := calculator.ProjectRent(history, calculator.RentProjectionInput{
		BaseAmount:                charge.BaseAmount,
		BasePeriod:                charge.BasePeriod,
		AdjustmentFrequencyMonths: charge.AdjustmentFrequencyMonths,
	})
	if err != nil {
		resp.ProjectionError = err.Error()
		return resp
	}
	resp.Projection = &result
	return resp
}

func (h recurringChargeHandlers) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req recurringChargeRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	charge, err := h.svc.Create(r.Context(), userID, recurringcharge.CreateInput{
		Name:                      req.Name,
		BaseAmount:                req.BaseAmount,
		BasePeriod:                req.BasePeriod,
		AdjustmentFrequencyMonths: req.AdjustmentFrequencyMonths,
		IndexCode:                 req.IndexCode,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, h.withProjection(r.Context(), charge))
}

func (h recurringChargeHandlers) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	charges, err := h.svc.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list recurring charges")
		return
	}
	out := make([]recurringChargeResponse, 0, len(charges))
	for _, c := range charges {
		out = append(out, h.withProjection(r.Context(), c))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h recurringChargeHandlers) get(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid recurring charge id")
		return
	}
	charge, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "recurring charge not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get recurring charge")
		return
	}
	writeJSON(w, http.StatusOK, h.withProjection(r.Context(), charge))
}

func (h recurringChargeHandlers) update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid recurring charge id")
		return
	}
	var req recurringChargeRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	charge, err := h.svc.Update(r.Context(), userID, id, recurringcharge.UpdateInput{
		Name:                      req.Name,
		BaseAmount:                req.BaseAmount,
		BasePeriod:                req.BasePeriod,
		AdjustmentFrequencyMonths: req.AdjustmentFrequencyMonths,
		IndexCode:                 req.IndexCode,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "recurring charge not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, h.withProjection(r.Context(), charge))
}

func (h recurringChargeHandlers) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid recurring charge id")
		return
	}
	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete recurring charge")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
