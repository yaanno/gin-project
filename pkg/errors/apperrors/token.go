package apperrors

// Token validation errors

type TokenError struct {
	baseError
}

func NewTokenError(code ErrorCode, message string, cause error) TokenError {
	return TokenError{
		baseError: baseError{
			message: message,
			code:    code,
			cause:   cause,
		},
	}
}

func (e TokenError) Code() ErrorCode { return e.code }
func (e TokenError) Error() string   { return e.message }
func (e TokenError) Unwrap() error   { return e.cause }
