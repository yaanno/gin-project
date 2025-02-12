// internal/services/user_service.go
package services

import (
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
		logger: logger.With().Str("service", "UserService").Logger(),
	}
}

func (s *UserServiceImpl) GetAllUsers() ([]database.User, error) {
	return s.repo.GetAllUsers()
}

func (s *UserServiceImpl) GetUserByID(userID uint) (*database.User, error) {
	return s.repo.FindUserByID(userID)
}

func (s *UserServiceImpl) UpdateUser(user *database.User) error {
	if err := s.validateUser(user); err != nil {
		return err
	}

	return s.repo.UpdateUser(user)
}

func (s *UserServiceImpl) DeleteUser(userID uint) error {
	return s.repo.DeleteUser(userID)
}

func (s *UserServiceImpl) validateUser(user *database.User) error {
	_, err := s.repo.FindUserByID(user.ID)
	if err != nil {
		return err
	}
	return nil
}
