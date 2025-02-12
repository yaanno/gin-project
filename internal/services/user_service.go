// internal/services/user_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"

	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/internal/database"
	"github.com/yourusername/user-management-api/internal/repository"
)

type UserServiceImpl struct {
	repo   repository.UserRepository
	logger zerolog.Logger
}

func NewUserService(repo repository.UserRepository, logger zerolog.Logger) *UserServiceImpl {
	return &UserServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

func (s *UserServiceImpl) Create(ctx context.Context, user *database.User) error {
	// Additional validation
	if err := s.validateUser(user); err != nil {
		return err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	// Log the user creation attempt
	s.logger.Info().
		Str("username", user.Username).
		Str("email", user.Email).
		Msg("Attempting to create user")

	return s.repo.CreateUser(user)
}

func (s *UserServiceImpl) Authenticate(ctx context.Context, email, password string) (*database.User, error) {
	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *UserServiceImpl) validateUser(user *database.User) error {
	// Implement complex validation logic
	if len(user.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	// Email validation
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(user.Email) {
		return errors.New("invalid email format")
	}

	return nil
}
