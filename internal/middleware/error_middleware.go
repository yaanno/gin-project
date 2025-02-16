package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/pkg/errors/apperrors"
)

type ErrorResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func ErrorMiddleware(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			handleErrors(c, log)
			return
		}
	}
}

func handleErrors(c *gin.Context, log zerolog.Logger) {
	// Get the first error from the context
	err := c.Errors.Last()
	if err == nil {
		return
	}

	// Log the error with comprehensive context
	logError(log, err, c)

	// Determine the appropriate response based on error type
	switch {
	case isAppError(err.Err):
		handleAppError(c, err.Err.(apperrors.AppError))
	case isValidationError(err):
		handleValidationError(c, err)
	default:
		handleGenericError(c, err)
	}
}

func logError(log zerolog.Logger, err *gin.Error, c *gin.Context) {
	logEntry := log.Error().
		Err(err).
		Str("uri", c.Request.URL.Path).
		Str("method", c.Request.Method)

	// Add additional context if it's an AppError
	if appErr, ok := err.Err.(apperrors.AppError); ok {
		logEntry = logEntry.
			Str("error_code", string(appErr.Code())).
			Str("error_message", appErr.Error())
	}

	logEntry.Msg("Request processing error")
}

func handleAppError(c *gin.Context, appErr apperrors.AppError) {
	status := mapAppErrorToHTTPStatus(appErr)

	response := ErrorResponse{
		Code:    status,
		Message: "An error occurred",
	}

	// Optionally include error details in non-production environments
	if gin.Mode() != gin.ReleaseMode {
		response.Details = map[string]interface{}{
			"error_code": appErr.Code(),
			"message":    appErr.Error(),
		}
	}

	c.JSON(status, response)
}

func handleValidationError(c *gin.Context, err *gin.Error) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: "Validation failed",
		Details: err.Error(),
	})
}

func handleGenericError(c *gin.Context, err *gin.Error) {
	status := http.StatusInternalServerError
	c.JSON(status, ErrorResponse{
		Code:    status,
		Message: "An unexpected error occurred",
		Details: err.Error(),
	})
}

func mapAppErrorToHTTPStatus(err apperrors.AppError) int {
	switch err.Code() {
	case apperrors.ErrCodeNotFound:
		return http.StatusNotFound
	case apperrors.ErrCodeInvalidCredentials:
		return http.StatusUnauthorized
	case apperrors.ErrCodeUserLocked, apperrors.ErrCodeUserInactive:
		return http.StatusForbidden
	case apperrors.ErrCodeDatabaseError:
		return http.StatusInternalServerError
	case apperrors.ErrCodeValidationError:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func isAppError(err error) bool {
	_, ok := err.(apperrors.AppError)
	return ok
}

func isValidationError(err *gin.Error) bool {
	return err.Type == gin.ErrorTypeBind
}
