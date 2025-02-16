// internal/services/interfaces.go
package services

import (
	"context"

	"github.com/yourusername/user-management-api/internal/database"
	"github.com/yourusername/user-management-api/pkg/errors/apperrors"
)

type UserService interface {
	GetAllUsers() ([]database.User, error)
	GetUserByID(userID uint) (*database.User, error)
	UpdateUser(user *database.User) error
	DeleteUser(userID uint) error
}

type AuthService interface {
	GenerateAccessToken(ctx context.Context, userID uint, username string) (string, apperrors.AppError)
	GenerateRefreshToken(ctx context.Context, userID uint, username string) (string, apperrors.AppError)
	ValidateRefreshToken(ctx context.Context, token string) (uint, string, error)
	// ValidateAccessToken(ctx context.Context, token string) (*database.User, apperrors.AppError)
	RefreshTokens(ctx context.Context, userID uint, username string) (*database.TokenPair, apperrors.AppError)

	RegisterUser(ctx context.Context, username, password, email string) (*database.User, error)
	LoginUser(ctx context.Context, username, password, ipAddr string) (*database.TokenPair, apperrors.AppError)
	LogoutUser(ctx context.Context, token string) error
}

type UserCleanupService interface {
	CleanupUsers() error
}

var _ AuthService = (*AuthServiceImpl)(nil)
var _ UserService = (*UserServiceImpl)(nil)
var _ UserCleanupService = (*UserCleanupServiceImpl)(nil)
