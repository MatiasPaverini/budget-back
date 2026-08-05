// Package authctx carries the authenticated user id through a request's
// context. It's deliberately independent of any specific auth method
// (Firebase, a dev bypass, whatever comes next) — every auth middleware
// populates it the same way, and every handler reads it the same way.
package authctx

import "context"

type contextKey string

const userIDKey contextKey = "userID"

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}
