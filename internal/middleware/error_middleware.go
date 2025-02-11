package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors in the context
		for _, err := range c.Errors {
			// Log the error
			log.Printf("Error: %v", err)

			// Determine the appropriate HTTP status code
			status := getStatusCode(err)

			// Respond with error details
			c.JSON(status, ErrorResponse{
				Code:    status,
				Message: err.Error(),
			})
			return
		}
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

func HandleNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, ErrorResponse{
		Code:    http.StatusNotFound,
		Message: "Endpoint not found",
	})
}
