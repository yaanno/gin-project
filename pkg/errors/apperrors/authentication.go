package apperrors

// AuthenticationError represents an authentication error.
type AuthenticationError struct {
	baseError
}

// NewAuthenticationError creates a new AuthenticationError.
func NewAuthenticationError(code ErrorCode, message string, cause error) AuthenticationError {
	return AuthenticationError{
		baseError: baseError{
			message: message,
			code:    code,
			cause:   cause,
		},
	}
}

// Code returns the error code for this error.
func (e AuthenticationError) Code() ErrorCode { return e.code }

// Error returns the error message for this error.
func (e AuthenticationError) Error() string { return e.message }

// Unwrap returns the underlying error for this error.
func (e AuthenticationError) Unwrap() error { return e.cause }
