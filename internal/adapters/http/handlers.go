package http

import (
	"context"

	"messaging-app/internal/domain"
)

const (
	UserContextKey = "user"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// GetUserFromContext extracts user context from request context
func GetUserFromContext(ctx context.Context) (domain.UserContext, bool) {
	user, ok := ctx.Value(UserContextKey).(domain.UserContext)
	return user, ok
}
