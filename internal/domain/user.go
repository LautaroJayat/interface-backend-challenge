package domain

import (
	"fmt"
	"strings"
)

type UserContext struct {
	UserID  string `json:"user_id" validate:"required,max=100"`
	Email   string `json:"email" validate:"required,email,max=255"`
	Handler string `json:"handler" validate:"required,max=50"`
}

// Validate performs user context validation
func (uc *UserContext) Validate() error {
	if strings.TrimSpace(uc.UserID) == "" {
		return ErrMissingUserID
	}
	if strings.TrimSpace(uc.Email) == "" {
		return ErrMissingEmail
	}
	if strings.TrimSpace(uc.Handler) == "" {
		return ErrMissingHandler
	}
	return nil
}

// String returns a string representation
func (uc *UserContext) String() string {
	return fmt.Sprintf("User{ID: %s, Email: %s, Handler: %s}", uc.UserID, uc.Email, uc.Handler)
}