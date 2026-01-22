package errors

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// ErrorCode represents a specific error type
type ErrorCode string

const (
	// Client errors (4xx)
	ErrCodeBadRequest       ErrorCode = "BAD_REQUEST"
	ErrCodeValidation       ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrCodeConflict         ErrorCode = "CONFLICT"
	ErrCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrCodeRateLimited      ErrorCode = "RATE_LIMITED"

	// Server errors (5xx)
	ErrCodeInternal         ErrorCode = "INTERNAL_ERROR"
	ErrCodeDatabaseError    ErrorCode = "DATABASE_ERROR"
	ErrCodeExchangeError    ErrorCode = "EXCHANGE_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"

	// Business logic errors
	ErrCodeJobLocked        ErrorCode = "JOB_LOCKED"
	ErrCodeConnectorInactive ErrorCode = "CONNECTOR_INACTIVE"
	ErrCodeSymbolInvalid    ErrorCode = "SYMBOL_INVALID"
	ErrCodeNoData           ErrorCode = "NO_DATA"
)

// APIError represents a structured API error
type APIError struct {
	Code       ErrorCode   `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	StatusCode int         `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   *APIError `json:"error"`
}

// NewAPIError creates a new API error
func NewAPIError(code ErrorCode, message string, statusCode int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WithDetails adds details to the error
func (e *APIError) WithDetails(details interface{}) *APIError {
	e.Details = details
	return e
}

// Common error constructors

// BadRequest creates a 400 Bad Request error
func BadRequest(message string) *APIError {
	return NewAPIError(ErrCodeBadRequest, message, fiber.StatusBadRequest)
}

// ValidationError creates a 400 Validation error
func ValidationError(message string, details interface{}) *APIError {
	return NewAPIError(ErrCodeValidation, message, fiber.StatusBadRequest).WithDetails(details)
}

// NotFound creates a 404 Not Found error
func NotFound(resource string) *APIError {
	return NewAPIError(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), fiber.StatusNotFound)
}

// Conflict creates a 409 Conflict error
func Conflict(message string) *APIError {
	return NewAPIError(ErrCodeConflict, message, fiber.StatusConflict)
}

// Unauthorized creates a 401 Unauthorized error
func Unauthorized(message string) *APIError {
	return NewAPIError(ErrCodeUnauthorized, message, fiber.StatusUnauthorized)
}

// Forbidden creates a 403 Forbidden error
func Forbidden(message string) *APIError {
	return NewAPIError(ErrCodeForbidden, message, fiber.StatusForbidden)
}

// RateLimited creates a 429 Rate Limited error
func RateLimited(message string) *APIError {
	return NewAPIError(ErrCodeRateLimited, message, fiber.StatusTooManyRequests)
}

// InternalError creates a 500 Internal Server Error
func InternalError(message string) *APIError {
	return NewAPIError(ErrCodeInternal, message, fiber.StatusInternalServerError)
}

// DatabaseError creates a 500 Database Error
func DatabaseError(message string) *APIError {
	return NewAPIError(ErrCodeDatabaseError, message, fiber.StatusInternalServerError)
}

// ExchangeError creates a 502 Exchange Error
func ExchangeError(message string) *APIError {
	return NewAPIError(ErrCodeExchangeError, message, fiber.StatusBadGateway)
}

// ServiceUnavailable creates a 503 Service Unavailable error
func ServiceUnavailable(message string) *APIError {
	return NewAPIError(ErrCodeServiceUnavailable, message, fiber.StatusServiceUnavailable)
}

// Business logic error constructors

// JobLocked creates a job locked error
func JobLocked(jobID string) *APIError {
	return NewAPIError(ErrCodeJobLocked, fmt.Sprintf("Job %s is currently locked by another process", jobID), fiber.StatusConflict)
}

// ConnectorInactive creates a connector inactive error
func ConnectorInactive(connectorID string) *APIError {
	return NewAPIError(ErrCodeConnectorInactive, fmt.Sprintf("Connector %s is not active", connectorID), fiber.StatusBadRequest)
}

// SymbolInvalid creates an invalid symbol error
func SymbolInvalid(symbol, exchange string) *APIError {
	return NewAPIError(ErrCodeSymbolInvalid, fmt.Sprintf("Symbol %s is not valid on exchange %s", symbol, exchange), fiber.StatusBadRequest)
}

// NoData creates a no data error
func NoData(resource string) *APIError {
	return NewAPIError(ErrCodeNoData, fmt.Sprintf("No data available for %s", resource), fiber.StatusNotFound)
}

// SendError sends an error response to the client
func SendError(c *fiber.Ctx, err *APIError) error {
	return c.Status(err.StatusCode).JSON(ErrorResponse{
		Success: false,
		Error:   err,
	})
}

// WrapError wraps a standard error into an APIError
func WrapError(err error, code ErrorCode, statusCode int) *APIError {
	return &APIError{
		Code:       code,
		Message:    err.Error(),
		StatusCode: statusCode,
	}
}
