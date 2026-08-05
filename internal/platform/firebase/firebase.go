package firebase

import (
	"context"
	"errors"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"

	"github.com/mpaverini/budget-back/internal/platform/authctx"
)

// Verifier checks Firebase ID tokens and identifies the calling user.
type Verifier struct {
	client *auth.Client
}

func NewVerifier(ctx context.Context, credentialsFile string) (*Verifier, error) {
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, err
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}
	return &Verifier{client: client}, nil
}

// Middleware verifies the Authorization: Bearer <token> header on every
// request and injects the Firebase UID into the request context via
// authctx. Every downstream handler scopes its queries by this UID.
func (v *Verifier) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := bearerToken(r)
		if err != nil {
			http.Error(w, "missing or malformed authorization header", http.StatusUnauthorized)
			return
		}

		decoded, err := v.client.VerifyIDToken(r.Context(), token)
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := authctx.WithUserID(r.Context(), decoded.UID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", errors.New("missing bearer token")
	}
	return strings.TrimPrefix(header, prefix), nil
}
