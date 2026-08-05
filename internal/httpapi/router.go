package httpapi

import (
	"net/http"

	"github.com/mpaverini/budget-back/internal/account"
	"github.com/mpaverini/budget-back/internal/indicator"
	"github.com/mpaverini/budget-back/internal/recurringcharge"
	"github.com/mpaverini/budget-back/internal/transaction"
)

type Dependencies struct {
	// Auth is whichever auth middleware is active (real Firebase
	// verification or the local dev bypass) — httpapi never needs to know
	// which.
	Auth             func(http.Handler) http.Handler
	Accounts         *account.Service
	Transactions     *transaction.Service
	Indicators       *indicator.Service
	RecurringCharges *recurringcharge.Service
	IPCSeriesID      string
}

func NewRouter(deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)

	api := http.NewServeMux()
	accountHandlers{deps.Accounts}.register(api)
	transactionHandlers{deps.Transactions}.register(api)
	indicatorHandlers{deps.Indicators, deps.IPCSeriesID}.register(api)
	calculatorHandlers{deps.Indicators}.register(api)
	recurringChargeHandlers{deps.RecurringCharges, deps.Indicators}.register(api)

	mux.Handle("/", deps.Auth(api))

	return withCORS(mux)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
