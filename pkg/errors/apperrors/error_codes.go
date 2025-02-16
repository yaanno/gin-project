package apperrors

// Define error codes
const (
	// Timeout and Rate Limit Errors
	ErrCodeTimeoutError           ErrorCode = "TIMEOUT_ERROR"
	ErrCodeCanceledError          ErrorCode = "CANCELED_ERROR"
	ErrCodeRateLimitError         ErrorCode = "RATE_LIMIT_ERROR"
	ErrCodeInvalidAPIKey          ErrorCode = "INVALID_API_KEY"
	ErrCodeInvalidRateLimitConfig ErrorCode = "INVALID_RATE_LIMIT_CONFIG"
	ErrCodeTooManyRequests        ErrorCode = "RATE_LIMIT_EXCEEDED"

	// Database Errors
	ErrCodeDatabaseError ErrorCode = "DATABASE_ERROR"

	// Not Found Errors
	ErrCodeNotFound ErrorCode = "NOT_FOUND"

	// Authentication Errors
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeUserLocked         ErrorCode = "USER_LOCKED"
	ErrCodeUserInactive       ErrorCode = "USER_INACTIVE"
	ErrCodeUserDeleted        ErrorCode = "USER_DELETED"

	// General Errors
	ErrCodeUnknownError  ErrorCode = "UNKNOWN_ERROR"
	ErrCodeInternalError ErrorCode = "INTERNAL_ERROR"
)
