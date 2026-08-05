package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/mpaverini/budget-back/internal/platform/authctx"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// requireUserID pulls the authenticated user id out of the request context
// (set by whichever auth middleware is wired up — see authctx) and writes a
// 401 if it's somehow missing, which should only happen if a route was
// mounted outside the auth middleware by mistake.
func requireUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID, ok := authctx.UserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing user")
		return "", false
	}
	return userID, true
}
