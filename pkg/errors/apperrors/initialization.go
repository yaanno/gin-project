package apperrors

// InitializationError represents an error during initialization
type InitializationError struct {
	baseError
}

// NewInitializationError creates a new InitializationError
func NewInitializationError(message string, cause error) InitializationError {
	return InitializationError{
		baseError: baseError{
			message: message,
			code:    ErrCodeInitializationError,
			cause:   cause,
		},
	}
}

// Error returns the error message for this error.
func (e InitializationError) Error() string { return e.message }

// Code returns the error code for this error.
func (e InitializationError) Code() ErrorCode { return e.code }

// Unwrap returns the underlying error for this error.
func (e InitializationError) Unwrap() error { return e.cause }
