package httpapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/mpaverini/budget-back/internal/account"
)

type accountHandlers struct {
	svc *account.Service
}

func (h accountHandlers) register(mux *http.ServeMux) {
	mux.HandleFunc("GET /accounts", h.list)
	mux.HandleFunc("POST /accounts", h.create)
	mux.HandleFunc("GET /accounts/{id}", h.get)
	mux.HandleFunc("PATCH /accounts/{id}", h.update)
	mux.HandleFunc("DELETE /accounts/{id}", h.delete)
}

type createAccountRequest struct {
	Name              string           `json:"name"`
	Type              account.Type     `json:"type"`
	Currency          string           `json:"currency"`
	OpeningBalance    decimal.Decimal  `json:"opening_balance"`
	OpenedAt          *time.Time       `json:"opened_at"`
	CreditLimit       *decimal.Decimal `json:"credit_limit"`
	StatementCloseDay *int16           `json:"statement_close_day"`
	DueDay            *int16           `json:"due_day"`
	InterestRate      *decimal.Decimal `json:"interest_rate"`
	TermMonths        *int16           `json:"term_months"`
}

func (h accountHandlers) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var req createAccountRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	in := account.CreateInput{
		Name:              req.Name,
		Type:              req.Type,
		Currency:          req.Currency,
		OpeningBalance:    req.OpeningBalance,
		CreditLimit:       req.CreditLimit,
		StatementCloseDay: req.StatementCloseDay,
		DueDay:            req.DueDay,
		InterestRate:      req.InterestRate,
		TermMonths:        req.TermMonths,
	}
	if req.OpenedAt != nil {
		in.OpenedAt = *req.OpenedAt
	}

	acc, err := h.svc.Create(r.Context(), userID, in)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, acc)
}

func (h accountHandlers) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	accounts, err := h.svc.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list accounts")
		return
	}
	writeJSON(w, http.StatusOK, accounts)
}

func (h accountHandlers) get(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}
	acc, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "account not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get account")
		return
	}
	writeJSON(w, http.StatusOK, acc)
}

type updateAccountRequest struct {
	Name              string           `json:"name"`
	Currency          string           `json:"currency"`
	CreditLimit       *decimal.Decimal `json:"credit_limit"`
	StatementCloseDay *int16           `json:"statement_close_day"`
	DueDay            *int16           `json:"due_day"`
	InterestRate      *decimal.Decimal `json:"interest_rate"`
	TermMonths        *int16           `json:"term_months"`
}

func (h accountHandlers) update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}
	var req updateAccountRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	acc, err := h.svc.Update(r.Context(), userID, id, account.UpdateInput{
		Name:              req.Name,
		Currency:          req.Currency,
		CreditLimit:       req.CreditLimit,
		StatementCloseDay: req.StatementCloseDay,
		DueDay:            req.DueDay,
		InterestRate:      req.InterestRate,
		TermMonths:        req.TermMonths,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "account not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, acc)
}

func (h accountHandlers) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}
	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete account")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
