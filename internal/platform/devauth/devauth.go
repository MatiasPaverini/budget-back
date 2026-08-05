// Package devauth is a stand-in for firebase.Verifier used only when
// AUTH_MODE=dev: it skips token verification entirely and injects a fixed
// user id into every request. It exists so the frontend can be built and
// exercised end to end before a real Firebase project is set up, without
// httpapi or any handler needing to know which auth method is active.
package devauth

import (
	"log"
	"net/http"

	"github.com/mpaverini/budget-back/internal/platform/authctx"
)

const DefaultUserID = "dev-user"

type Middleware struct {
	UserID string
}

func NewMiddleware(userID string) Middleware {
	if userID == "" {
		userID = DefaultUserID
	}
	log.Printf("WARNING: AUTH_MODE=dev — every request is trusted as user %q, no token is checked. Do not use this outside local development.", userID)
	return Middleware{UserID: userID}
}

func (m Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authctx.WithUserID(r.Context(), m.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
