package testclient

import "fmt"

// GetMessagesOptions holds options for message retrieval
type GetMessagesOptions struct {
	Cursor string // RFC3339 timestamp for pagination
	Limit  int    // Maximum number of messages to retrieve
}

// APIError represents an API error response
type APIError struct {
	StatusCode int
	Message    string
	Code       string
	Details    string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("API error %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// Error checking helper functions

func IsStatusCode(err error, statusCode int) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == statusCode
	}
	return false
}

func IsUnauthorized(err error) bool {
	return IsStatusCode(err, 401)
}

func IsForbidden(err error) bool {
	return IsStatusCode(err, 403)
}

func IsBadRequest(err error) bool {
	return IsStatusCode(err, 400)
}

func IsErrorCode(err error, errorCode string) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code == errorCode
	}
	return false
}

// Error codes from the API
const (
	ErrorCodeMissingUserID    = "MISSING_USER_ID"
	ErrorCodeAccessDenied     = "ACCESS_DENIED"
	ErrorCodeValidationError  = "VALIDATION_ERROR"
)