package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/internal/database"
	"github.com/yourusername/user-management-api/internal/repository"
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

type TokenBlacklist struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

func (tb *TokenBlacklist) Add(token string, expiration time.Time) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.tokens[token] = expiration
}

func (tb *TokenBlacklist) IsBlacklisted(token string) bool {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	// Remove expired tokens
	for t, exp := range tb.tokens {
		if time.Now().After(exp) {
			delete(tb.tokens, t)
		}
	}

	_, exists := tb.tokens[token]
	return exists
}

var GlobalTokenBlacklist = &TokenBlacklist{
	tokens: make(map[string]time.Time),
}

type AuthHandlerImpl struct {
	repo   repository.UserRepository
	logger zerolog.Logger
}

func NewAuthHandler(repo repository.UserRepository, logger zerolog.Logger) *AuthHandlerImpl {
	return &AuthHandlerImpl{
		repo:   repo,
		logger: logger,
	}
}

func (a *AuthHandlerImpl) RegisterUser(c *gin.Context) {
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

	user := &database.User{
		Username: req.Username,
		Email:    req.Email,
		Password: sanitizedPassword,
	}

	// Hash password before storing
	if err := user.HashPassword(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Save user to database
	if err := a.repo.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func (a *AuthHandlerImpl) LoginUser(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by username
	user, err := a.repo.FindUserByUsername(req.Username)
	if err != nil {
		a.logger.Err(err).Msg("User not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Check password
	if !user.CheckPasswordHash(req.Password) {
		a.logger.Err(err).Msg("Invalid credentials")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate token pair
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		a.logger.Err(err).Msg("Failed to generate access token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		a.logger.Err(err).Msg("Failed to generate refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	c.JSON(http.StatusOK, database.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (a *AuthHandlerImpl) RefreshTokens(c *gin.Context) {
	var refreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&refreshRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate refresh token
	claims, err := utils.ValidateToken(refreshRequest.RefreshToken, "refresh")
	if err != nil {
		a.logger.Err(err).Msg("Invalid refresh token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Generate new token pair
	accessToken, err := utils.GenerateAccessToken(claims.UserID, claims.Username)
	if err != nil {
		a.logger.Err(err).Msg("Failed to generate access token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(claims.UserID, claims.Username)
	if err != nil {
		a.logger.Err(err).Msg("Failed to generate refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	c.JSON(http.StatusOK, database.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (a *AuthHandlerImpl) LogoutUser(c *gin.Context) {
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

	// Validate token to get claims
	claims, err := utils.ValidateToken(token, "access")
	if err != nil {
		a.logger.Err(err).Msg("Invalid token during logout")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Calculate token expiration (use token's expiration or a default)
	expiration := time.Now().Add(24 * time.Hour)

	// Blacklist the token
	GlobalTokenBlacklist.Add(token, expiration)

	a.logger.Info().
		Str("user_id", strconv.Itoa(int(claims.UserID))).
		Msg("User logged out successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

func CheckTokenBlacklist(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}
		token := parts[1]

		// Check if token is blacklisted
		if GlobalTokenBlacklist.IsBlacklisted(token) {
			logger.Info().Msg("Blacklisted token used")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is no longer valid"})
			c.Abort()
			return
		}

		c.Next()
	}
}

var _ AuthHandler = &AuthHandlerImpl{}
