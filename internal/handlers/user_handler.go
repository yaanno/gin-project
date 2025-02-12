package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/internal/services"
	"github.com/yourusername/user-management-api/pkg/utils"
)

type UserHandlerImpl struct {
	service services.UserService
	log     zerolog.Logger
}

func NewUserHandler(userService services.UserService, log zerolog.Logger) *UserHandlerImpl {
	return &UserHandlerImpl{
		service: userService,
		log:     log.With().Str("handler", "UserHandler").Logger(),
	}
}

func (h *UserHandlerImpl) GetAllUsers(c *gin.Context) {
	_, cancel := utils.GetContextWithTimeout()
	defer cancel()
	users, err := h.service.GetAllUsers()
	if err != nil {
		h.log.Err(err).Msg("Failed to get all users")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get all users"})
		return
	}

	// Mask sensitive information
	var safeUsers []SafeUser
	for _, user := range users {
		safeUsers = append(safeUsers, SafeUser{ // Use SafeUser struct here
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format(time.RFC3339), // Format time.Time if needed here or in SafeUser struct
		})
	}

	c.JSON(http.StatusOK, GetAllUsersResponse{Users: safeUsers})
}

func (h *UserHandlerImpl) GetUserByID(c *gin.Context) {
	_, cancel := utils.GetContextWithTimeout()
	defer cancel()
	// Get user ID from path parameter
	userID, ok := h.parseUserID(c)
	if !ok {
		return // parseUserID already handled error response
	}

	// Retrieve user from database
	user, err := h.service.GetUserByID(uint(userID))
	if err != nil {
		h.log.Err(err).Msg("Failed to get user by ID")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get user by ID", Details: err.Error()})
		return
	}

	// Return safe user information
	c.JSON(http.StatusOK, GetUserByIDResponse{User: SafeUser{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}})
}

func (h *UserHandlerImpl) UpdateUser(c *gin.Context) {
	_, cancel := utils.GetContextWithTimeout()
	defer cancel()
	// Get user ID from path parameter
	userID, ok := h.parseUserID(c)
	if !ok {
		return // parseUserID already handled error response
	}

	// Bind update request
	var updateReq struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"omitempty"`
	}
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body", Details: err.Error()})
		return
	}

	// Find existing user
	user, err := h.service.GetUserByID(uint(userID))
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to get user by ID")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get user by ID", Details: err.Error()})
		return
	}

	// Update email
	user.Email = updateReq.Email

	// Update password if provided
	if updateReq.Password != "" {
		// Validate password complexity
		sanitizedPassword := utils.SanitizePassword(updateReq.Password)
		if !utils.IsPasswordComplex(sanitizedPassword) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Password does not meet complexity requirements"})
			return
		}

		// Hash new password
		user.Password = sanitizedPassword
		if err := user.HashPassword(); err != nil {
			h.log.Error().Err(err).Msg("Failed to hash password")
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to hash password", Details: err.Error()})
			return
		}
	}

	// Save updated user
	if err := h.service.UpdateUser(user); err != nil {
		h.log.Error().Err(err).Msg("Failed to update user")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update user", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, UpdateUserResponse{Message: "User updated successfully"})
}

func (h *UserHandlerImpl) DeleteUser(c *gin.Context) {
	_, cancel := utils.GetContextWithTimeout()
	defer cancel()
	// Get user ID from path parameter
	userID, ok := h.parseUserID(c)
	if !ok {
		return // parseUserID already handled error response
	}

	// Delete user
	if err := h.service.DeleteUser(uint(userID)); err != nil {
		h.log.Error().Err(err).Msg("Failed to delete user")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete user", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, DeleteUserResponse{Message: "User deleted successfully"})
}

func (h *UserHandlerImpl) parseUserID(c *gin.Context) (uint64, bool) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		h.log.Error().Err(err).Msg("Invalid user ID")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID", Details: err.Error()})
		return 0, false
	}
	return userID, true
}
