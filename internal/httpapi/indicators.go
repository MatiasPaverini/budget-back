package httpapi

import (
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	"github.com/mpaverini/budget-back/internal/indicator"
)

type indicatorHandlers struct {
	svc         *indicator.Service
	ipcSeriesID string
}

func (h indicatorHandlers) register(mux *http.ServeMux) {
	mux.HandleFunc("GET /indicators/ipc", h.history)
	mux.HandleFunc("POST /indicators/ipc", h.recordManual)
	mux.HandleFunc("POST /indicators/ipc/sync", h.sync)
}

func (h indicatorHandlers) history(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireUserID(w, r); !ok {
		return
	}
	points, err := h.svc.History(r.Context(), indicator.IPCCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load IPC history")
		return
	}
	writeJSON(w, http.StatusOK, points)
}

type recordIndicatorRequest struct {
	Period time.Time       `json:"period"`
	Value  decimal.Decimal `json:"value"`
}

func (h indicatorHandlers) recordManual(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireUserID(w, r); !ok {
		return
	}
	var req recordIndicatorRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	point, err := h.svc.RecordManual(r.Context(), indicator.IPCCode, req.Period, req.Value)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, point)
}

func (h indicatorHandlers) sync(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireUserID(w, r); !ok {
		return
	}
	count, err := h.svc.SyncIPC(r.Context(), h.ipcSeriesID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"synced": count})
}
