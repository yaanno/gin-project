// internal/services/interfaces.go
package services

import (
	"time"

	"github.com/yourusername/user-management-api/internal/database"
)

type UserService interface {
	GetAllUsers() ([]database.User, error)
	GetUserByID(userID uint) (*database.User, error)
	UpdateUser(user *database.User) error
	DeleteUser(userID uint) error
}

type AuthService interface {
	GenerateAccessToken(userID uint, username string) (string, error)
	GenerateRefreshToken(userID uint, username string) (string, error)
	ValidateRefreshToken(token string) (*database.User, error)
	ValidateAccessToken(token string) (*database.User, error)
	RefreshTokens(userID uint, username string) (*database.TokenPair, error)

	RegisterUser(username, password string) (*database.User, error)
	LoginUser(username, password string) (*database.TokenPair, error)
	LogoutUser(token string) error

	IsTokenBlacklisted(token string) bool
	AddTokenToBlacklist(token string, expiration time.Time)
}

var _ AuthService = (*AuthServiceImpl)(nil)
var _ UserService = (*UserServiceImpl)(nil)
