package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/user-management-api/internal/repository"
	"github.com/yourusername/user-management-api/pkg/utils"
)

type UserHandler interface {
	GetAllUsers(c *gin.Context)
}

type UserHandlerImpl struct {
	repo repository.UserRepository
}

func NewUserHandler(repo repository.UserRepository) *UserHandlerImpl {
	return &UserHandlerImpl{
		repo: repo,
	}
}

func (h *UserHandlerImpl) GetAllUsers(c *gin.Context) {
	users, err := h.repo.GetAllUsers()
	if err != nil {
		c.Error(err)
		return
	}

	// Mask sensitive information
	var safeUsers []gin.H
	for _, user := range users {
		safeUsers = append(safeUsers, gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, safeUsers)
}

func (h *UserHandlerImpl) GetUserByID(c *gin.Context) {
	// Get user ID from path parameter
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Retrieve user from database
	user, err := h.repo.FindUserByID(uint(userID))
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.Error(err)
		return
	}

	// Return safe user information
	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt,
	})
}

func (h *UserHandlerImpl) UpdateUser(c *gin.Context) {
	// Get user ID from path parameter
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Bind update request
	var updateReq struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"omitempty"`
	}
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find existing user
	user, err := h.repo.FindUserByID(uint(userID))
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.Error(err)
		return
	}

	// Update email
	user.Email = updateReq.Email

	// Update password if provided
	if updateReq.Password != "" {
		// Validate password complexity
		sanitizedPassword := utils.SanitizePassword(updateReq.Password)
		if !utils.IsPasswordComplex(sanitizedPassword) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Password does not meet complexity requirements",
			})
			return
		}

		// Hash new password
		user.Password = sanitizedPassword
		if err := user.HashPassword(); err != nil {
			c.Error(err)
			return
		}
	}

	// Save updated user
	if err := h.repo.UpdateUser(user); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func (h *UserHandlerImpl) DeleteUser(c *gin.Context) {
	// Get user ID from path parameter
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Delete user
	if err := h.repo.DeleteUser(uint(userID)); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
