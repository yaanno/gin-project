// internal/services/auth_service.go
package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/internal/database"
	"github.com/yourusername/user-management-api/internal/repository"
	"github.com/yourusername/user-management-api/pkg/utils"
)

type TokenBlacklist struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

var GlobalTokenBlacklist = &TokenBlacklist{
	tokens: make(map[string]time.Time),
}

type AuthServiceImpl struct {
	logger    zerolog.Logger
	repo      repository.UserRepository
	blacklist *TokenBlacklist
}

func NewAuthService(repo repository.UserRepository, logger zerolog.Logger) *AuthServiceImpl {
	return &AuthServiceImpl{
		repo:      repo,
		logger:    logger,
		blacklist: GlobalTokenBlacklist,
	}
}

func (s *AuthServiceImpl) GenerateAccessToken(userID uint, username string) (string, error) {
	return utils.GenerateAccessToken(userID, username)
}

func (s *AuthServiceImpl) GenerateRefreshToken(userID uint, username string) (string, error) {
	return utils.GenerateRefreshToken(userID, username)
}

func (s *AuthServiceImpl) RefreshTokens(userID uint, username string) (*database.TokenPair, error) {
	accessToken, err := s.GenerateAccessToken(userID, username)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.GenerateRefreshToken(userID, username)
	if err != nil {
		return nil, err
	}

	return &database.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthServiceImpl) ValidateAccessToken(tokenString string) (*database.User, error) {
	claims, err := utils.ValidateToken(tokenString, "access")
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

func (s *AuthServiceImpl) ValidateRefreshToken(tokenString string) (*database.User, error) {
	claims, err := utils.ValidateToken(tokenString, "refresh")
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type")
	}

	user, err := s.repo.FindUserByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *TokenBlacklist) Add(token string, expiration time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token] = expiration
}

func (s *TokenBlacklist) IsBlacklisted(token string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Remove expired tokens
	for t, exp := range s.tokens {
		if time.Now().After(exp) {
			delete(s.tokens, t)
		}
	}

	_, exists := s.tokens[token]
	return exists
}

func (s *AuthServiceImpl) AddTokenToBlacklist(token string, expiration time.Time) {
	s.blacklist.Add(token, expiration)
}

func (s *AuthServiceImpl) IsTokenBlacklisted(token string) bool {
	return s.blacklist.IsBlacklisted(token)
}

func (s *AuthServiceImpl) LoginUser(username, password string) (*database.TokenPair, error) {
	// Find user by username
	user, err := s.repo.FindUserByUsername(username)
	if err != nil {
		return nil, err
	}

	// Check password
	if !user.CheckPasswordHash(password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	tokens, err := s.RefreshTokens(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthServiceImpl) LogoutUser(token string) error {
	if s.IsTokenBlacklisted(token) {
		return fmt.Errorf("token is blacklisted")
	}

	claims, err := utils.ValidateToken(token, "access")
	if err != nil {
		return fmt.Errorf("invalid token")
	}

	if claims.TokenType != "access" {
		return fmt.Errorf("invalid token type")
	}

	s.AddTokenToBlacklist(token, time.Now().Add(24*time.Hour))

	return nil
}

func (s *AuthServiceImpl) RegisterUser(username, password string) (*database.User, error) {

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
