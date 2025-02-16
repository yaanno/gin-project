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

	// Validation Errors
	ErrCodeValidationError ErrorCode = "VALIDATION_ERROR"

	// Initialization Errors
	ErrCodeInitializationError ErrorCode = "INITIALIZATION_ERROR"

	// Token validation errors
	ErrCodeInvalidToken          ErrorCode = "INVALID_TOKEN"
	ErrCodeTokenExpired          ErrorCode = "TOKEN_EXPIRED"
	ErrCodeTokenMalformed        ErrorCode = "TOKEN_MALFORMED"
	ErrCodeTokenInvalidClaim     ErrorCode = "TOKEN_INVALID_CLAIM"
	ErrCodeTokenBlacklisted      ErrorCode = "TOKEN_BLACKLISTED"
	ErrCodeTokenInvalidType      ErrorCode = "TOKEN_INVALID_TYPE"
	ErrCodeInvalidTokenSignature ErrorCode = "INVALID_TOKEN_SIGNATURE"
	ErrCodeTokenSigningError     ErrorCode = "TOKEN_SIGNING_ERROR"
	ErrCodeParseError            ErrorCode = "PARSE_ERROR"

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
