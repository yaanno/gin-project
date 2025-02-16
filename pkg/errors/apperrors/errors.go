package apperrors

type ErrorCode string

type AppError interface {
	Error() string
	Code() ErrorCode
	Unwrap() error
	Is(target error, code ErrorCode) bool
}

type baseError struct {
	message string
	code    ErrorCode
	cause   error
}

func (e baseError) Error() string   { return e.message }
func (e baseError) Code() ErrorCode { return e.code }
func (e baseError) Unwrap() error   { return e.cause }

func (e baseError) Is(target error, code ErrorCode) bool {
	if target == nil {
		return false
	}

	appErr, ok := target.(AppError)
	if !ok {
		return false
	}

	return appErr.Code() == code
}
