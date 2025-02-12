package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e ErrorResponse) Error() string {
	return e.Message
}

func ErrorMiddleware(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Check if there are any errors in the context
		for _, err := range c.Errors {
			// Determine the appropriate HTTP status code
			status := getStatusCode(err)
			log.Error().
				Err(err).
				Str("status", strconv.Itoa(status)).
				Str("uri", c.Request.URL.Path).
				Str("method", c.Request.Method).
				Msg("Error handling request")
			// Respond with error details
			c.JSON(status, ErrorResponse{
				Code:    status,
				Message: err.Error(),
			})
			return
		}

		// If no errors, continue to the next middleware
		c.Next()
	}
}

func getStatusCode(err *gin.Error) int {
	switch err.Type {
	case gin.ErrorTypeBind:
		return http.StatusBadRequest
	case gin.ErrorTypeRender:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func HandleNotFound(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Warn().
			Str("uri", c.Request.URL.Path).
			Str("method", c.Request.Method).
			Msg("Endpoint not found")
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "Endpoint not found",
		})
	}
}
