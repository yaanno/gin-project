package apperrors

// InternalError represents an internal error.
type InternalError struct {
	baseError
}

// NewInternalError creates a new InternalError.
func NewInternalError(message string, cause error) InternalError {
	return InternalError{
		baseError: baseError{
			message: message,
			code:    ErrCodeInternalError,
			cause:   cause,
		},
	}
}

// Error returns the error message for this error.
func (e InternalError) Error() string { return e.message }

// Code returns the error code for this error.
func (e InternalError) Code() ErrorCode { return e.code }

// Unwrap returns the underlying error for this error.
func (e InternalError) Unwrap() error { return e.cause }
