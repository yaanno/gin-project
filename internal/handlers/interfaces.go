package handlers

import "github.com/gin-gonic/gin"

// Define response structs
type GetAllUsersResponse struct {
	Users []SafeUser `json:"users"` // Assuming you'll define SafeUser struct
}

type GetUserByIDResponse struct {
	User SafeUser `json:"user"`
}

type SafeUser struct { // Struct to represent safe user info
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`           // Consider time.Time and format in handler if needed
	UpdatedAt string `json:"updated_at,omitempty"` //omitempty if not always present
}

type UpdateUserResponse struct { // If you want to return something specific on update success
	Message string `json:"message"`
}

type DeleteUserResponse struct { // If you want to return something specific on delete success
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"` // Optional details for debugging (hide in prod)
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserHandler interface {
	GetAllUsers(c *gin.Context)
	GetUserByID(c *gin.Context)
	UpdateUser(c *gin.Context)
	DeleteUser(c *gin.Context)
}

type AuthHandler interface {
	RegisterUser(c *gin.Context)
	LoginUser(c *gin.Context)
	RefreshTokens(c *gin.Context)
	LogoutUser(c *gin.Context)
}

var _ UserHandler = (*UserHandlerImpl)(nil)
var _ AuthHandler = (*AuthHandlerImpl)(nil)
