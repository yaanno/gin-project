package apperrors

// Validation error
type ValidationErrors struct {
	baseError
}

// NewValidationErrors creates a new ValidationErrors
func NewValidationErrors(message string, cause error) ValidationErrors {
	return ValidationErrors{
		baseError: baseError{
			message: message,
			code:    ErrCodeValidationError,
			cause:   cause,
		},
	}
}

func (e ValidationErrors) Unwrap() error   { return e.cause }
func (e ValidationErrors) Error() string   { return e.message }
func (e ValidationErrors) Code() ErrorCode { return e.code }
