package apperrors

// Not found error
type NotFoundError struct {
	baseError
	resourceType string
	identifier   interface{}
}

// NewNotFoundError creates a new Not found error
func NewNotFoundError(message string, cause error, resourceType string, identifier interface{}) NotFoundError {
	return NotFoundError{
		baseError: baseError{
			message: message,
			code:    ErrCodeNotFound,
			cause:   cause,
		},
		resourceType: resourceType,
		identifier:   identifier,
	}
}

// Error returns the error message for this error.
func (e NotFoundError) Error() string { return e.message }

// Code returns the error code for this error.
func (e NotFoundError) Code() ErrorCode { return e.code }

// Unwrap returns the underlying error for this error.
func (e NotFoundError) Unwrap() error { return e.cause }
