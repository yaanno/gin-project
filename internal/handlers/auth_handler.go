package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

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
}

type AuthHandlerImpl struct {
	repo repository.UserRepository
}

func NewAuthHandler(repo repository.UserRepository) *AuthHandlerImpl {
	return &AuthHandlerImpl{
		repo: repo,
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if !user.CheckPasswordHash(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate token pair
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Generate new token pair
	accessToken, err := utils.GenerateAccessToken(claims.UserID, claims.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(claims.UserID, claims.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	c.JSON(http.StatusOK, database.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
