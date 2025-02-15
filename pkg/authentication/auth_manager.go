package authentication

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"math/rand"

	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/internal/database"
	"github.com/yourusername/user-management-api/internal/repository"
	"github.com/yourusername/user-management-api/pkg/token"
)

const MAX_LOGIN_ATTEMPTS = 5

type LoginAttempt struct {
	Username    string
	IPAddress   string
	Attempts    int
	LastAttempt time.Time
	Locked      bool
	LockUntil   time.Time
}

type AuthenticationManager struct {
	userRepo         repository.UserRepository
	tokenManager     *token.TokenManager
	loginAttemptRepo repository.LoginAttemptRepository
	logger           zerolog.Logger
}

func NewAuthenticationManager(
	userRepo repository.UserRepository,
	tokenManager *token.TokenManager,
	loginAttemptRepo repository.LoginAttemptRepository,
	logger zerolog.Logger,
) *AuthenticationManager {
	return &AuthenticationManager{
		userRepo:         userRepo,
		tokenManager:     tokenManager,
		loginAttemptRepo: loginAttemptRepo,
		logger:           logger,
	}
}

func (am *AuthenticationManager) ValidateUserAuthentication(
	ctx context.Context,
	username string,
	password string,
	ipAddress string,
) (*database.User, error) {
	// Consolidated validation logic
	user, err := am.FindUserByUsername(username)
	if err != nil {
		return nil, err
	}

	// 1. Check User Status
	if err := am.CheckUserStatus(user); err != nil {
		return nil, err
	}

	// 2. Validate Password
	if !user.CheckPasswordHash(password) {
		// Record failed login attempt
		am.recordFailedLoginAttempt(username, ipAddress)
		return nil, errors.New("invalid credentials")
	}

	// 3. Check Login Attempts
	if err := am.checkLoginAttempts(username, user.ID, ipAddress); err != nil {
		return nil, err
	}

	// Reset successful login attempts
	am.resetLoginAttempts(username, ipAddress)

	return user, nil
}

func (am *AuthenticationManager) FindUserByUsername(username string) (*database.User, error) {
	return am.userRepo.FindUserByUsername(username)
}

func (am *AuthenticationManager) CheckUserStatus(user *database.User) error {
	switch user.Status {
	case database.UserStatusLocked:
		if user.LockedUntil.After(time.Now()) {
			return fmt.Errorf("account locked until %s. Reason: %s",
				user.LockedUntil.Format(time.RFC3339),
				user.LockReason)
		}
	case database.UserStatusInactive, database.UserStatusDeleted:
		return errors.New("account is not active")
	}
	return nil
}

func (am *AuthenticationManager) calculateLockDelay(attempts int) time.Duration {
	// Exponential backoff with randomization
	baseDelay := 1 * time.Second
	maxDelay := 1 * time.Hour
	// Exponential calculation
	delay := baseDelay * time.Duration(math.Pow(2, float64(attempts-1)))
	// Randomization
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
	delay += jitter

	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

func (am *AuthenticationManager) checkLoginAttempts(
	username string,
	userID uint,
	ipAddress string,
) error {
	attempts, _, err := am.loginAttemptRepo.GetLoginAttempts(username, ipAddress)
	if err != nil {
		return err
	}

	if attempts >= MAX_LOGIN_ATTEMPTS {
		lockDuration := am.calculateLockDelay(attempts)
		// Automatically lock the user
		err := am.userRepo.LockUser(
			userID,
			"Exceeded maximum login attempts",
			lockDuration,
		)
		if err != nil {
			am.logger.Error().Err(err).Msg("Failed to lock user")
		}
		return errors.New("too many login attempts. Account locked")
	}

	return nil
}

func (am *AuthenticationManager) recordFailedLoginAttempt(
	username string,
	ipAddress string,
) {
	attempts, _, _ := am.loginAttemptRepo.GetLoginAttempts(username, ipAddress)
	possibleLockDuration := am.calculateLockDelay(attempts + 1)

	am.logger.Warn().
		Str("username", username).
		Str("ipAddress", ipAddress).
		Int("attempts", attempts+1).
		Dur("next_attempt_delay", possibleLockDuration).
		Msg("Progressive login delay applied")

	// Record the failed login attempt
	err := am.loginAttemptRepo.IncrementLoginAttempts(username, ipAddress, false)
	if err != nil {
		am.logger.Error().Err(err).Msg("Failed to record login attempt")
	}
}

func (am *AuthenticationManager) resetLoginAttempts(
	username string,
	ipAddress string,
) {
	err := am.loginAttemptRepo.ResetLoginAttempts(username, ipAddress)
	if err != nil {
		am.logger.Error().Err(err).Msg("Failed to reset login attempts")
	}
}

func (am *AuthenticationManager) ValidateToken(tokenString string, tokenType token.TokenType) (*token.Claims, error) {
	claims, err := am.tokenManager.ValidateToken(tokenString, tokenType)
	if err != nil {
		return nil, err
	}

	return claims, nil

}
