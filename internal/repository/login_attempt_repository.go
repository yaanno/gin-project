package repository

import (
	"database/sql"
	"time"

	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/internal/database/sqlite"
)

type LoginAttemptRepository interface {
	IncrementLoginAttempts(username string, ipAddress string, success bool) error
	ResetLoginAttempts(username string, ipAddress string) error
	GetLoginAttempts(username string, ipAddress string) (int, time.Time, error)
}

type LoginAttemptRepositoryImpl struct {
	db  *sqlite.SQLiteDatabase
	log zerolog.Logger
}

func NewLoginAttemptRepository(db *sqlite.SQLiteDatabase, log zerolog.Logger) *LoginAttemptRepositoryImpl {
	return &LoginAttemptRepositoryImpl{
		db:  db,
		log: log.With().Str("repository", "LoginAttemptRepository").Logger(),
	}
}

func (r *LoginAttemptRepositoryImpl) IncrementLoginAttempts(username string, ipAddress string, success bool) error {
	query := `
        INSERT OR IGNORE INTO login_attempts (username, ip_address, attempts, last_attempt, success)
        VALUES (?, ?, 0, ?, ?) ON CONFLICT(username, ip_address) DO UPDATE SET attempts = attempts + 1, last_attempt = ?, success = ?
    `

	_, err := r.db.ExecuteQuery(query, username, ipAddress, time.Now(), success, time.Now(), success)
	if err != nil {
		r.log.Error().Err(err).Msg("Failed to record login attempt")
		return err
	}

	return nil
}

func (r *LoginAttemptRepositoryImpl) ResetLoginAttempts(username string, ipAddress string) error {
	_, err := r.db.ExecuteQuery("UPDATE login_attempts SET attempts = 0, last_attempt = NULL, success = FALSE WHERE username = ? AND ip_address = ?", username, ipAddress)
	if err != nil {
		r.log.Error().Err(err).Msg("Failed to reset login attempts")
		return err
	}
	return nil
}

func (r *LoginAttemptRepositoryImpl) GetLoginAttempts(username string, ipAddress string) (int, time.Time, error) {
	var attempts int
	var lastAttempt time.Time
	err := r.db.QueryRow("SELECT attempts, last_attempt FROM login_attempts WHERE username = ? AND ip_address = ?", username, ipAddress).Scan(&attempts, &lastAttempt)
	if err == sql.ErrNoRows {
		return 0, time.Time{}, nil // No attempts recorded yet
	} else if err != nil {
		r.log.Error().Err(err).Msg("Failed to get login attempts")
		return 0, time.Time{}, err
	}
	return attempts, lastAttempt, nil
}

var _ LoginAttemptRepository = (*LoginAttemptRepositoryImpl)(nil)
