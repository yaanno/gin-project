package apperrors

type DatabaseError struct {
	baseError
}

func NewDatabaseError(message string, cause error) DatabaseError {
	return DatabaseError{
		baseError: baseError{
			message: message,
			code:    ErrCodeDatabaseError,
			cause:   cause,
		},
	}
}

func (e DatabaseError) Error() string   { return e.message }
func (e DatabaseError) Code() ErrorCode { return e.code }
func (e DatabaseError) Unwrap() error   { return e.cause }
