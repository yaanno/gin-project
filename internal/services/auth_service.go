// internal/services/auth_service.go
package services

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/internal/database"
	"github.com/yourusername/user-management-api/internal/repository"
	"github.com/yourusername/user-management-api/pkg/token"
	"github.com/yourusername/user-management-api/pkg/utils"
)

type AuthServiceImpl struct {
	logger       zerolog.Logger
	repo         repository.UserRepository
	tokenManager *token.TokenManager
}

func NewAuthService(tokenManager *token.TokenManager, repo repository.UserRepository, logger zerolog.Logger) *AuthServiceImpl {
	return &AuthServiceImpl{
		repo:         repo,
		logger:       logger.With().Str("service", "AuthService").Logger(),
		tokenManager: tokenManager,
	}
}

func (s *AuthServiceImpl) GenerateAccessToken(ctx context.Context, userID uint, username string) (string, error) {
	return s.tokenManager.GenerateToken(userID, username, "access")
}

func (s *AuthServiceImpl) GenerateRefreshToken(ctx context.Context, userID uint, username string) (string, error) {
	return s.tokenManager.GenerateToken(userID, username, "refresh")
}

func (s *AuthServiceImpl) RefreshTokens(ctx context.Context, userID uint, username string) (*database.TokenPair, error) {

	accessToken, err := s.GenerateAccessToken(ctx, userID, username)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.GenerateRefreshToken(ctx, userID, username)
	if err != nil {
		return nil, err
	}

	return &database.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthServiceImpl) ValidateAccessToken(ctx context.Context, tokenString string) (*database.User, error) {
	claims, err := s.tokenManager.ValidateToken(tokenString, "access")
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, fmt.Errorf("invalid token type")
	}

	user, err := s.repo.FindUserByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthServiceImpl) ValidateRefreshToken(ctx context.Context, tokenString string) (uint, string, error) {
	claims, err := s.tokenManager.ValidateToken(tokenString, "refresh")
	if err != nil {
		return 0, "", err
	}

	if claims.TokenType != "refresh" {
		return 0, "", fmt.Errorf("invalid token type")
	}
	return claims.UserID, claims.Username, nil
}

func (s *AuthServiceImpl) LoginUser(ctx context.Context, username, password string) (*database.TokenPair, error) {
	// Find user by username
	user, err := s.repo.FindUserByUsername(username)
	if err != nil {
		return nil, err
	}

	// Check password
	if !user.CheckPasswordHash(password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	tokens, err := s.RefreshTokens(ctx, user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthServiceImpl) LogoutUser(ctx context.Context, token string) error {
	if err := s.tokenManager.InvalidateToken(token); err != nil {
		return err
	}
	return nil
}

func (s *AuthServiceImpl) RegisterUser(ctx context.Context, username, password string) (*database.User, error) {
	_, cancel := utils.GetContextWithTimeout()
	defer cancel()

	user := &database.User{
		Username: username,
		Password: password,
	}

	if err := user.HashPassword(); err != nil {
		return nil, err
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}
