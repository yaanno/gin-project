package services

import (
	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/internal/repository"
)

type UserCleanupServiceImpl struct {
	userRepo repository.UserRepository
	log      zerolog.Logger
}

func NewUserCleanupService(userRepo repository.UserRepository, log zerolog.Logger) *UserCleanupServiceImpl {
	return &UserCleanupServiceImpl{
		userRepo: userRepo,
		log:      log.With().Str("service", "UserCleanupService").Logger(),
	}
}

func (s *UserCleanupServiceImpl) CleanupUsers() error {
	// Lock users with repeated security violations
	err := s.lockSecurityViolationUsers()
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to lock security violation users")
	}

	// Mark long-inactive users as inactive
	err = s.markInactiveUsers()
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to mark inactive users")
	}

	// Hard delete permanently inactive users
	err = s.hardDeletePermanentlyInactiveUsers()
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to hard delete inactive users")
	}

	return nil
}

func (s *UserCleanupServiceImpl) lockSecurityViolationUsers() error {
	if err := s.userRepo.LockSecurityViolationUsers(); err != nil {
		return err
	}
	return nil
}

func (s *UserCleanupServiceImpl) markInactiveUsers() error {
	if err := s.userRepo.MarkInactiveUsers(); err != nil {
		return err
	}
	return nil
}

func (s *UserCleanupServiceImpl) hardDeletePermanentlyInactiveUsers() error {
	if err := s.userRepo.HardDeletePermanentlyInactiveUsers(); err != nil {
		return err
	}
	return nil
}
