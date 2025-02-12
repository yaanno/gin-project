package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/internal/services"
	"github.com/yourusername/user-management-api/pkg/utils"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthHandler interface {
	RegisterUser(c *gin.Context)
	LoginUser(c *gin.Context)
	RefreshTokens(c *gin.Context)
	LogoutUser(c *gin.Context)
}

type AuthHandlerImpl struct {
	service *services.AuthServiceImpl
	logger  zerolog.Logger
}

func NewAuthHandler(authService *services.AuthServiceImpl, logger zerolog.Logger) *AuthHandlerImpl {
	return &AuthHandlerImpl{
		service: authService,
		logger:  logger,
	}
}

func (a *AuthHandlerImpl) RegisterUser(c *gin.Context) {
	ctx, cancel := utils.GetContextWithTimeout()
	defer cancel()
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate password complexity
	sanitizedPassword := utils.SanitizePassword(req.Password)
	if !utils.IsPasswordComplex(sanitizedPassword) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password does not meet complexity requirements. " +
				"Minimum 12 characters with uppercase, lowercase, number, and special character.",
		})
		return
	}

	user, err := a.service.RegisterUser(ctx, req.Username, sanitizedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user": user})
}

func (a *AuthHandlerImpl) LoginUser(c *gin.Context) {
	ctx, cancel := utils.GetContextWithTimeout()
	defer cancel()
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenPair, err := a.service.LoginUser(ctx, req.Username, req.Password)
	if err != nil {
		a.logger.Err(err).Msg("Failed to login user")
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate refresh token
	accessToken, err := a.service.ValidateRefreshToken(ctx, refreshRequest.RefreshToken)
	if err != nil {
		a.logger.Err(err).Msg("Invalid refresh token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	tokenPair, err := a.service.RefreshTokens(ctx, accessToken.ID, accessToken.Username)
	if err != nil {
		a.logger.Err(err).Msg("Failed to refresh tokens")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to refresh tokens"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header missing"})
		return
	}

	// Extract token (expecting "Bearer <token>")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		a.logger.Info().Msg("Invalid authorization format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization format"})
		return
	}
	token := parts[1]

	if err := a.service.LogoutUser(ctx, token); err != nil {
		a.logger.Err(err).Msg("Failed to logout user")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to logout user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User logged out successfully"})
}

var _ AuthHandler = &AuthHandlerImpl{}
