package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/internal/services"
	"github.com/yourusername/user-management-api/pkg/utils"
)

type AuthHandlerImpl struct {
	service *services.AuthServiceImpl
	logger  zerolog.Logger
}

func NewAuthHandler(authService *services.AuthServiceImpl, logger zerolog.Logger) *AuthHandlerImpl {
	return &AuthHandlerImpl{
		service: authService,
		logger:  logger.With().Str("handler", "AuthHandler").Logger(),
	}
}

func (a *AuthHandlerImpl) RegisterUser(c *gin.Context) {
	ctx, cancel := utils.GetContextWithTimeout()
	defer cancel()
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body", Details: err.Error()})
		return
	}

	// Validate password complexity
	sanitizedPassword := utils.SanitizePassword(req.Password)
	if !utils.IsPasswordComplex(sanitizedPassword) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Password does not meet complexity requirements. " +
			"Minimum 12 characters with uppercase, lowercase, number, and special character."})
		return
	}

	user, err := a.service.RegisterUser(ctx, req.Username, sanitizedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user": user})
}

func (a *AuthHandlerImpl) LoginUser(c *gin.Context) {
	ctx, cancel := utils.GetContextWithTimeout()
	defer cancel()
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body", Details: err.Error()})
		return
	}

	tokenPair, err := a.service.LoginUser(ctx, req.Username, req.Password)
	if err != nil {
		a.logger.Err(err).Msg("Failed to login user")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokenPair)
}

func (a *AuthHandlerImpl) RefreshTokens(c *gin.Context) {
	ctx, cancel := utils.GetContextWithTimeout()
	defer cancel()
	var refreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&refreshRequest); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body", Details: err.Error()})
		return
	}

	// Validate refresh token
	userID, username, err := a.service.ValidateRefreshToken(ctx, refreshRequest.RefreshToken)
	if err != nil {
		a.logger.Err(err).Msg("Invalid refresh token")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid refresh token"})
		return
	}

	tokenPair, err := a.service.RefreshTokens(ctx, userID, username)
	if err != nil {
		a.logger.Err(err).Msg("Failed to refresh tokens")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Failed to refresh tokens", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokenPair)
}

func (a *AuthHandlerImpl) LogoutUser(c *gin.Context) {
	ctx, cancel := utils.GetContextWithTimeout()
	defer cancel()
	// Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		a.logger.Info().Msg("No authorization header")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Authorization header missing"})
		return
	}

	// Extract token (expecting "Bearer <token>")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		a.logger.Info().Msg("Invalid authorization format")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid authorization format"})
		return
	}
	token := parts[1]

	if err := a.service.LogoutUser(ctx, token); err != nil {
		a.logger.Err(err).Msg("Failed to logout user")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Failed to logout user", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User logged out successfully"})
}
