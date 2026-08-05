package httpapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/mpaverini/budget-back/internal/transaction"
)

type transactionHandlers struct {
	svc *transaction.Service
}

func (h transactionHandlers) register(mux *http.ServeMux) {
	mux.HandleFunc("GET /transactions", h.list)
	mux.HandleFunc("POST /transactions", h.create)
	mux.HandleFunc("PATCH /transactions/{id}", h.update)
	mux.HandleFunc("DELETE /transactions/{id}", h.delete)
	mux.HandleFunc("GET /networth", h.netWorth)
}

type createTransactionRequest struct {
	AccountID         uuid.UUID       `json:"account_id"`
	OccurredAt        *time.Time      `json:"occurred_at"`
	Amount            decimal.Decimal `json:"amount"`
	Description       string          `json:"description"`
	Category          *string         `json:"category"`
	TransferAccountID *uuid.UUID      `json:"transfer_account_id"`
}

func (h transactionHandlers) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req createTransactionRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in := transaction.CreateInput{
		AccountID:         req.AccountID,
		Amount:            req.Amount,
		Description:       req.Description,
		Category:          req.Category,
		TransferAccountID: req.TransferAccountID,
	}
	if req.OccurredAt != nil {
		in.OccurredAt = *req.OccurredAt
	}
	tx, err := h.svc.Create(r.Context(), userID, in)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, tx)
}

func (h transactionHandlers) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var filter transaction.ListFilter
	if raw := r.URL.Query().Get("account_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid account_id")
			return
		}
		filter.AccountID = &id
	}
	if raw := r.URL.Query().Get("from"); raw != "" {
		t, err := time.Parse("2006-01-02", raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid from date, expected YYYY-MM-DD")
			return
		}
		filter.From = &t
	}
	if raw := r.URL.Query().Get("to"); raw != "" {
		t, err := time.Parse("2006-01-02", raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid to date, expected YYYY-MM-DD")
			return
		}
		filter.To = &t
	}

	txs, err := h.svc.List(r.Context(), userID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list transactions")
		return
	}
	writeJSON(w, http.StatusOK, txs)
}

type updateTransactionRequest struct {
	OccurredAt  *time.Time      `json:"occurred_at"`
	Amount      decimal.Decimal `json:"amount"`
	Description string          `json:"description"`
	Category    *string         `json:"category"`
}

func (h transactionHandlers) update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}
	var req updateTransactionRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in := transaction.UpdateInput{
		Amount:      req.Amount,
		Description: req.Description,
		Category:    req.Category,
	}
	if req.OccurredAt != nil {
		in.OccurredAt = *req.OccurredAt
	}
	tx, err := h.svc.Update(r.Context(), userID, id, in)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "transaction not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tx)
}

func (h transactionHandlers) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}
	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete transaction")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h transactionHandlers) netWorth(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	balances, err := h.svc.NetWorth(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to compute net worth")
		return
	}

	total := decimal.Zero
	for _, b := range balances {
		total = total.Add(b.Balance)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"accounts": balances,
		"total":    total,
	})
}
