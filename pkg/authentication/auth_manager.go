package authentication

import (
	"context"
	"fmt"
	"time"

	"math/rand"

	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/internal/database"
	"github.com/yourusername/user-management-api/internal/repository"
	"github.com/yourusername/user-management-api/pkg/errors/apperrors"
	"github.com/yourusername/user-management-api/pkg/token"
)

const MAX_LOGIN_ATTEMPTS = 5

type AuthenticationManager interface {
	CheckUserStatus(user *database.User) error
	CalculateLockDelay(attempts int) time.Duration
	CheckLoginAttempts(
		username string,
		userID uint,
		ipAddress string,
	) error
	ValidateToken(tokenString string, tokenType token.TokenType) (*token.Claims, error)
}

type AuthenticationManagerImpl struct {
	userRepo         repository.UserRepository
	tokenManager     token.TokenManager
	loginAttemptRepo repository.LoginAttemptRepository
	logger           zerolog.Logger
}

func NewAuthenticationManager(
	userRepo repository.UserRepository,
	tokenManager token.TokenManager,
	loginAttemptRepo repository.LoginAttemptRepository,
	logger zerolog.Logger,
) *AuthenticationManagerImpl {
	return &AuthenticationManagerImpl{
		userRepo:         userRepo,
		tokenManager:     tokenManager,
		loginAttemptRepo: loginAttemptRepo,
		logger:           logger,
	}
}

func (am *AuthenticationManagerImpl) ValidateUserAuthentication(
	ctx context.Context,
	username string,
	password string,
	ipAddress string,
) (*database.User, error) {
	// Consolidated validation logic
	user, err := am.FindUserByUsername(username)
	if err != nil {
		return nil, apperrors.NewNotFoundError("User not found", err, "user", username)
	}

	// 1. Check User Status
	if err := am.CheckUserStatus(user); err != nil {
		return nil, apperrors.NewInternalError("Failed to check user status", err)
	}

	// 2. Validate Password
	if !user.CheckPasswordHash(password) {
		// Record failed login attempt
		if err := am.recordFailedLoginAttempt(username, ipAddress); err != nil {
			return nil, apperrors.NewInternalError("Failed to record login attempt", err)
		}
		return nil, apperrors.NewAuthenticationError(apperrors.ErrCodeInvalidCredentials, "invalid credentials", nil)
	}

	// 3. Check Login Attempts
	if err := am.CheckLoginAttempts(username, user.ID, ipAddress); err != nil {
		return nil, err
	}

	// Reset successful login attempts
	am.resetLoginAttempts(username, ipAddress)

	return user, nil
}

func (am *AuthenticationManagerImpl) FindUserByUsername(username string) (*database.User, error) {
	return am.userRepo.FindUserByUsername(username)
}

func (am *AuthenticationManagerImpl) CheckUserStatus(user *database.User) error {
	switch user.Status {
	case database.UserStatusLocked:
		if user.LockedUntil.After(time.Now()) {
			return apperrors.NewAuthenticationError(apperrors.ErrCodeUserLocked, fmt.Sprintf("account locked until %s. Reason: %s",
				user.LockedUntil.Format(time.RFC3339),
				user.LockReason), nil)
		}
	case database.UserStatusInactive, database.UserStatusDeleted:
		return apperrors.NewAuthenticationError(apperrors.ErrCodeUserInactive, "account inactive", nil)
	}
	return nil
}

func (am *AuthenticationManagerImpl) CalculateLockDelay(attempts int) time.Duration {
	// Exponential backoff with randomization
	maxDelay := 1 * time.Hour

	// Ensure attempts starts from 1
	if attempts < 1 {
		attempts = 1
	}

	// Predefined delay steps to ensure precise control
	delaySteps := []time.Duration{
		1 * time.Second,  // 1st attempt
		2 * time.Second,  // 2nd attempt
		4 * time.Second,  // 3rd attempt
		8 * time.Second,  // 4th attempt
		16 * time.Second, // 5th attempt
		32 * time.Second, // 6th attempt
		1 * time.Minute,  // 7th attempt
		2 * time.Minute,  // 8th attempt
		4 * time.Minute,  // 9th attempt
		1 * time.Hour,    // 10th and subsequent attempts
	}

	// Select the appropriate delay step
	var rawDelay time.Duration
	if attempts-1 < len(delaySteps) {
		rawDelay = delaySteps[attempts-1]
	} else {
		rawDelay = maxDelay
	}

	// Randomization (limited to 10% of the current delay)
	jitterMax := rawDelay / 10
	jitter := time.Duration(rand.Intn(int(jitterMax.Milliseconds()))) * time.Millisecond
	delay := rawDelay + jitter

	// Final cap to ensure we don't exceed maxDelay
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

func (am *AuthenticationManagerImpl) CheckLoginAttempts(
	username string,
	userID uint,
	ipAddress string,
) error {
	attempts, _, err := am.loginAttemptRepo.GetLoginAttempts(username, ipAddress)
	if err != nil {
		return err
	}

	if attempts >= MAX_LOGIN_ATTEMPTS {
		lockDuration := am.CalculateLockDelay(attempts)
		// Automatically lock the user
		err := am.userRepo.LockUser(
			userID,
			"Exceeded maximum login attempts",
			lockDuration,
		)
		if err != nil {
			am.logger.Error().Err(err).Msg("Failed to lock user")
			return apperrors.NewInternalError("Failed to lock user", err)
		}
		return apperrors.NewAuthenticationError(apperrors.ErrCodeUserLocked, "too many login attempts. Account locked", nil)
	}

	return nil
}

func (am *AuthenticationManagerImpl) recordFailedLoginAttempt(
	username string,
	ipAddress string,
) error {
	attempts, _, _ := am.loginAttemptRepo.GetLoginAttempts(username, ipAddress)
	possibleLockDuration := am.CalculateLockDelay(attempts + 1)

	am.logger.Warn().
		Str("username", username).
		Str("ipAddress", ipAddress).
		Int("attempts", int(attempts)+1).
		Dur("next_attempt_delay", possibleLockDuration).
		Msg("Progressive login delay applied")

	// Record the failed login attempt
	err := am.loginAttemptRepo.IncrementLoginAttempts(username, ipAddress, false)
	if err != nil {
		am.logger.Error().Err(err).Msg("Failed to record login attempt")
		return apperrors.NewInternalError("Failed to record login attempt", err)
	}

	return nil
}

func (am *AuthenticationManagerImpl) resetLoginAttempts(
	username string,
	ipAddress string,
) error {
	if err := am.loginAttemptRepo.ResetLoginAttempts(username, ipAddress); err != nil {
		am.logger.Error().Err(err).Msg("Failed to reset login attempts")
		return apperrors.NewInternalError("Failed to reset login attempts", err)
	}
	return nil
}

func (am *AuthenticationManagerImpl) ValidateToken(tokenString string, tokenType token.TokenType) (*token.Claims, error) {
	claims, err := am.tokenManager.ValidateToken(tokenString, tokenType)
	if err != nil {
		return nil, err
	}

	return claims, nil

}

var _ AuthenticationManager = (*AuthenticationManagerImpl)(nil)
