package errors

import (
	"fmt"
	"net/http"
)

// AppError represents a custom application error
type AppError struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	Internal   error       `json:"-"`
	RequestID  string      `json:"request_id,omitempty"`
	ErrorType  string      `json:"error_type"`
}

func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

// NewAppError creates a new AppError
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Internal:  err,
		ErrorType: "application_error",
	}
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: "Validation failed",
		Details: []ValidationError{{Field: field, Message: message}},
		ErrorType: "validation_error",
	}
}

// IsNotFound checks if error is a not found error
func IsNotFound(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == http.StatusNotFound
	}
	return false
}

// IsValidationError checks if error is a validation error
func IsValidationError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.ErrorType == "validation_error"
	}
	return false
}
