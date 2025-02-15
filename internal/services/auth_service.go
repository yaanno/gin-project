// internal/services/auth_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/internal/database"
	"github.com/yourusername/user-management-api/internal/repository"
	"github.com/yourusername/user-management-api/pkg/token"
	"github.com/yourusername/user-management-api/pkg/utils"
)

const MAX_LOGIN_ATTEMPTS = 5

type AuthServiceImpl struct {
	logger           zerolog.Logger
	repo             repository.UserRepository
	loginAttemptRepo repository.LoginAttemptRepository
	tokenManager     *token.TokenManager
}

func NewAuthService(tokenManager *token.TokenManager, repo repository.UserRepository, loginAttemptRepo repository.LoginAttemptRepository, logger zerolog.Logger) *AuthServiceImpl {
	return &AuthServiceImpl{
		repo:             repo,
		logger:           logger.With().Str("service", "AuthService").Logger(),
		loginAttemptRepo: loginAttemptRepo,
		tokenManager:     tokenManager,
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
		return &database.TokenPair{}, err
	}
	refreshToken, err := s.GenerateRefreshToken(ctx, userID, username)
	if err != nil {
		return &database.TokenPair{}, err
	}

	return &database.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// TODO: CURRENTLY UNUSED, IMPLEMENT LATER
func (s *AuthServiceImpl) ValidateAccessToken(ctx context.Context, tokenString string) (*database.User, error) {
	claims, err := s.tokenManager.ValidateToken(tokenString, "access")
	if err != nil {
		return &database.User{}, err
	}

	if claims.TokenType != "access" {
		return &database.User{}, fmt.Errorf("invalid token type")
	}

	user, err := s.repo.FindUserByID(claims.UserID)
	if err != nil {
		return &database.User{}, err
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

func (s *AuthServiceImpl) LoginUser(ctx context.Context, username, password, ipAddr string) (*database.TokenPair, error) {
	// check for lockout
	attempts, err := s.loginAttemptRepo.GetLoginAttempts(username, ipAddr)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get login attempts")
		return nil, fmt.Errorf("internal server error")
	}

	if attempts >= MAX_LOGIN_ATTEMPTS {
		s.logger.Warn().Msgf("Too many failed login attempts for user: %s from IP: %s", username, ipAddr)
		return nil, fmt.Errorf("too many failed login attempts")
	}

	// Find user by username
	user, err := s.repo.FindUserByUsername(username)
	if err != nil {
		// Log the error for debugging, but don't expose details to the user
		s.logger.Error().Err(err).Msgf("Failed to find user: %s", username)

		// Increment failed attempts even if the user isn't found (for brute-force protection)
		if err := s.loginAttemptRepo.IncrementLoginAttempts(username, ipAddr, false); err != nil {
			s.logger.Error().Err(err).Msg("Failed to record login attempt") // Log this, but don't stop the flow
		}
		return &database.TokenPair{}, err
	}

	// Check if user is locked
	switch user.Status {
	case database.UserStatusLocked:
		if user.LockedUntil.After(time.Now()) {
			return &database.TokenPair{}, fmt.Errorf("user is locked until %s, reason: %s", user.LockedUntil.Format(time.RFC3339), user.LockReason)
		}
	case database.UserStatusInactive, database.UserStatusDeleted:
		return &database.TokenPair{}, fmt.Errorf("user is not active")
	}

	// Check password
	if !user.CheckPasswordHash(password) {
		s.logger.Info().Msgf("Failed login attempt for user: %s from IP: %s", username, ipAddr) // Log the failed attempt

		if err := s.loginAttemptRepo.IncrementLoginAttempts(username, ipAddr, false); err != nil {
			s.logger.Error().Err(err).Msg("Failed to record login attempt")
		}
		return &database.TokenPair{}, fmt.Errorf("invalid credentials")
	}

	tokens, err := s.RefreshTokens(ctx, user.ID, user.Username)
	if err != nil {
		return &database.TokenPair{}, err
	}

	return tokens, nil
}

func (s *AuthServiceImpl) LogoutUser(ctx context.Context, token string) error {
	if err := s.tokenManager.InvalidateToken(token); err != nil {
		return err
	}
	return nil
}

func (s *AuthServiceImpl) RegisterUser(ctx context.Context, username, password, email string) (*database.User, error) {
	_, cancel := utils.GetContextWithTimeout()
	defer cancel()

	user := &database.User{
		Username: username,
		Password: password,
		Email:    email,
	}

	if err := user.HashPassword(); err != nil {
		return &database.User{}, err
	}

	if err := s.repo.CreateUser(user); err != nil {
		return &database.User{}, err
	}

	return user, nil
}
