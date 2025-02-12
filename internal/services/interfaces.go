// internal/services/interfaces.go
package services

import (
	"context"
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
	GenerateAccessToken(ctx context.Context, userID uint, username string) (string, error)
	GenerateRefreshToken(ctx context.Context, userID uint, username string) (string, error)
	ValidateRefreshToken(ctx context.Context, token string) (*database.User, error)
	ValidateAccessToken(ctx context.Context, token string) (*database.User, error)
	RefreshTokens(ctx context.Context, userID uint, username string) (*database.TokenPair, error)

	RegisterUser(ctx context.Context, username, password string) (*database.User, error)
	LoginUser(ctx context.Context, username, password string) (*database.TokenPair, error)
	LogoutUser(ctx context.Context, token string) error

	IsTokenBlacklisted(token string) bool
	AddTokenToBlacklist(token string, expiration time.Time)
}

var _ AuthService = (*AuthServiceImpl)(nil)
var _ UserService = (*UserServiceImpl)(nil)
