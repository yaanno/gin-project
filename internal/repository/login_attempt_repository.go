package repository

import (
	"time"

	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/internal/database"
	"gorm.io/gorm"
)

type LoginAttemptRepository interface {
	IncrementLoginAttempts(username string, ipAddress string, success bool) error
	ResetLoginAttempts(username string, ipAddress string) error
	GetLoginAttempts(username string, ipAddress string) (uint, time.Time, error)
}

type LoginAttemptRepositoryImpl struct {
	db  *gorm.DB
	log zerolog.Logger
}

func NewLoginAttemptRepository(db *gorm.DB, log zerolog.Logger) *LoginAttemptRepositoryImpl {
	return &LoginAttemptRepositoryImpl{
		db:  db,
		log: log.With().Str("repository", "LoginAttemptRepository").Logger(),
	}
}

func (r *LoginAttemptRepositoryImpl) IncrementLoginAttempts(username string, ipAddress string, success bool) error {

	newLoginAttempt := database.LoginAttempt{
		Username:    username,
		IpAddress:   ipAddress,
		Attempts:    1,
		Success:     success,
		LastAttempt: time.Now(),
	}

	var existingAttempt database.LoginAttempt

	r.log.Info().Str("username", username).Str("ip_address", ipAddress).Msg("Incrementing login attempts")

	result := r.db.Where(database.LoginAttempt{
		Username:  username,
		IpAddress: ipAddress,
	}).First(&existingAttempt)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		r.log.Error().
			Err(result.Error).
			Str("username", username).
			Str("ip_address", ipAddress).
			Msg("Failed to check existing login attempts")
		return result.Error
	}

	// If record doesn't exist, create a new one
	if result.Error == gorm.ErrRecordNotFound {
		createResult := r.db.Create(newLoginAttempt)
		if createResult.Error != nil {
			r.log.Error().
				Err(createResult.Error).
				Str("username", username).
				Str("ip_address", ipAddress).
				Msg("Failed to create login attempt")
			return createResult.Error
		}
		return nil
	}

	// If record exists, increment attempts
	updateResult := r.db.Model(&existingAttempt).
		Where("username = ? AND ip_address = ?", username, ipAddress).
		Updates(map[string]interface{}{
			"attempts":     gorm.Expr("attempts + 1"),
			"success":      success,
			"last_attempt": time.Now(),
		})

	if updateResult.Error != nil {
		r.log.Error().
			Err(updateResult.Error).
			Str("username", username).
			Str("ip_address", ipAddress).
			Msg("Failed to increment login attempts")
		return updateResult.Error
	}

	r.log.Info().
		Str("username", username).
		Str("ip_address", ipAddress).
		Int("attempts", int(existingAttempt.Attempts)+1).
		Msg("Login attempts incremented")

	return nil
}

func (r *LoginAttemptRepositoryImpl) ResetLoginAttempts(username string, ipAddress string) error {
	r.log.Info().Str("username", username).Str("ip_address", ipAddress).Msg("Resetting login attempts")
	loginAttempt := &database.LoginAttempt{
		Username:  username,
		IpAddress: ipAddress,
	}
	result := r.db.Where(database.LoginAttempt{
		Username:  username,
		IpAddress: ipAddress,
	}).First(loginAttempt)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		r.log.Error().Err(result.Error).Str("username", username).Str("ip_address", ipAddress).Msg("Failed to reset login attempts")
		return result.Error
	}

	// If no record exists, return
	if result.Error == gorm.ErrRecordNotFound {
		r.log.Info().Str("username", username).Str("ip_address", ipAddress).Msg("No login attempts to reset")
		return nil
	}

	// Update the record
	r.log.Info().Str("username", username).Str("ip_address", ipAddress).Msg("Resetting login attempts")
	updated := r.db.Model(loginAttempt).
		Where("username = ? AND ip_address = ?", username, ipAddress).
		Updates(map[string]interface{}{
			"attempts":     0,
			"success":      false,
			"last_attempt": nil,
		})

	if updated.Error != nil {
		r.log.Error().Err(updated.Error).Str("username", username).Str("ip_address", ipAddress).Msg("Failed to reset login attempts")
		return updated.Error
	}

	r.log.Info().Str("username", username).Str("ip_address", ipAddress).Msg("Login attempts reset")
	return nil
}

func (r *LoginAttemptRepositoryImpl) GetLoginAttempts(username string, ipAddress string) (uint, time.Time, error) {
	var loginAttempt database.LoginAttempt

	result := r.db.Where(database.LoginAttempt{
		Username:  username,
		IpAddress: ipAddress,
	}).First(&loginAttempt)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		r.log.Error().Err(result.Error).Str("username", username).Str("ip_address", ipAddress).Msg("Failed to get login attempts")
		return 0, time.Time{}, result.Error
	}

	if result.Error == gorm.ErrRecordNotFound {
		r.log.Info().Str("username", username).Str("ip_address", ipAddress).Msg("No login attempts found")
		return 0, time.Time{}, nil
	}

	r.log.Info().Str("username", username).Str("ip_address", ipAddress).Msg("Login attempts found")
	return loginAttempt.Attempts, loginAttempt.LastAttempt, nil
}

var _ LoginAttemptRepository = (*LoginAttemptRepositoryImpl)(nil)
