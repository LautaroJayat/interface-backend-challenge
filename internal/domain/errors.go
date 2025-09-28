package domain

import "errors"

// Domain errors
var (
	ErrInvalidSenderID   = errors.New("invalid sender ID")
	ErrInvalidReceiverID = errors.New("invalid receiver ID")
	ErrSelfMessage       = errors.New("cannot send message to self")
	ErrEmptyContent      = errors.New("message content cannot be empty")
	ErrContentTooLong    = errors.New("message content exceeds maximum length")
	ErrInvalidStatus     = errors.New("invalid message status")
	ErrMissingUserID     = errors.New("user ID is required")
	ErrMissingEmail      = errors.New("user email is required")
	ErrMissingHandler    = errors.New("user handler is required")
	ErrChatNotFound      = errors.New("chat not found")
	ErrMessageNotFound   = errors.New("message not found")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrDuplicateMessage  = errors.New("duplicate message")
)

// IsValidationError checks if error is domain validation related
func IsValidationError(err error) bool {
	validationErrors := []error{
		ErrInvalidSenderID, ErrInvalidReceiverID, ErrSelfMessage,
		ErrEmptyContent, ErrContentTooLong, ErrInvalidStatus,
		ErrMissingUserID, ErrMissingEmail, ErrMissingHandler,
	}

	for _, ve := range validationErrors {
		if errors.Is(err, ve) {
			return true
		}
	}
	return false
}