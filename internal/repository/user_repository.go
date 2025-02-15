package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"github.com/yourusername/user-management-api/internal/database"
	"github.com/yourusername/user-management-api/internal/database/sqlite"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRepository interface {
	CreateUser(user *database.User) error
	FindUserByUsername(username string) (*database.User, error)
	FindUserByID(userID uint) (*database.User, error)
	UpdateUser(user *database.User) error
	DeleteUser(userID uint) error
	GetAllUsers() ([]database.User, error)
	LockUser(userID uint, reason string, duration time.Duration) error
	UnlockUser(userID uint) error
	MarkUserInactive(userID uint) error
	HardDeleteUser(userID uint) error
	HardDeletePermanentlyInactiveUsers() error
	LockSecurityViolationUsers() error
	MarkInactiveUsers() error
}

type UserRepositoryImpl struct {
	db  *sqlite.SQLiteDatabase
	log zerolog.Logger
}

func NewUserRepository(db *sqlite.SQLiteDatabase, log zerolog.Logger) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		db:  db,
		log: log.With().Str("repository", "UserRepository").Logger(),
	}
}

func (r *UserRepositoryImpl) CreateUser(user *database.User) error {
	query := `
		INSERT INTO users (username, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, datetime('now'), datetime('now'))
		RETURNING id
	`
	err := r.db.QueryRow(query, user.Username, user.Email, user.Password).Scan(&user.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			r.log.Error().Err(err).Msg("User not found")
			return ErrUserNotFound
		}
		r.log.Error().Err(err).Msg("Failed to create user")
		return err
	}
	return nil
}

func (r *UserRepositoryImpl) FindUserByUsername(username string) (*database.User, error) {
	user := &database.User{}
	query := `
		SELECT id, username, email, password 
		FROM users 
		WHERE username = $1
	`
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
	)

	if err == sql.ErrNoRows {
		r.log.Error().Err(err).Msg("User not found")
		return &database.User{}, ErrUserNotFound
	}

	if err != nil {
		r.log.Error().Err(err).Msg("Failed to find user")
		return &database.User{}, err
	}

	return user, nil
}

func (r *UserRepositoryImpl) FindUserByID(userID uint) (*database.User, error) {
	user := &database.User{}
	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users 
		WHERE id = $1
	`
	err := r.db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		r.log.Error().Err(err).Msg("User not found")
		return &database.User{}, ErrUserNotFound
	}

	if err != nil {
		r.log.Error().Err(err).Msg("Failed to find user")
		return &database.User{}, err
	}

	return user, nil
}

func (r *UserRepositoryImpl) UpdateUser(user *database.User) error {
	query := `
		UPDATE users 
		SET email = $1, password = $2, updated_at = datetime('now')
		WHERE id = $3
	`
	_, err := r.db.ExecuteQuery(query, user.Email, user.Password, user.ID)
	if err != nil {
		r.log.Error().Err(err).Msg("Failed to update user")
		return err
	}
	return nil
}

func (r *UserRepositoryImpl) DeleteUser(userID uint) error {
	query := `
        UPDATE users 
        SET deleted_at = datetime('now')
        WHERE id = $1 AND deleted_at IS NULL
    `
	result, err := r.db.ExecuteQuery(query, userID)
	if err != nil {
		r.log.Error().Err(err).Msg("Failed to delete user")
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepositoryImpl) GetAllUsers() ([]database.User, error) {
	query := `
		SELECT id, username, email, created_at, updated_at 
		FROM users 
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		r.log.Error().Err(err).Msg("Failed to get users")
		return []database.User{}, err
	}
	defer rows.Close()

	var users []database.User
	for rows.Next() {
		var user database.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			r.log.Error().Err(err).Msg("Failed to get users")
			return []database.User{}, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return []database.User{}, err
	}

	return users, nil
}

func (r *UserRepositoryImpl) LockUser(userID uint, reason string, duration time.Duration) error {
	lockedUntil := time.Now().Add(duration)
	query := `
		UPDATE users 
		SET 
			status = 'locked', 
			lock_reason = $2, 
			locked_until = $3
		WHERE id = $1 AND status != 'deleted'
	`
	result, err := r.db.ExecuteQuery(query, userID, reason, lockedUntil)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepositoryImpl) UnlockUser(userID uint) error {
	query := `
		UPDATE users 
		SET 
			status = 'active', 
			lock_reason = NULL, 
			locked_until = NULL
		WHERE id = $1
	`
	_, err := r.db.ExecuteQuery(query, userID)
	return err
}

func (r *UserRepositoryImpl) MarkUserInactive(userID uint) error {
	query := `
		UPDATE users 
		SET 
			status = 'inactive', 
			deleted_at = datetime('now')
		WHERE id = $1 AND status != 'deleted'
	`
	_, err := r.db.ExecuteQuery(query, userID)
	return err
}

func (r *UserRepositoryImpl) HardDeleteUser(userID uint) error {
	// Optional: Implement data archiving before hard delete
	archiveQuery := `
		INSERT INTO user_archive 
		SELECT * FROM users WHERE id = $1
	`

	deleteQuery := `DELETE FROM users WHERE id = $1 AND status = 'deleted'`

	tx, err := r.db.BeginTx(context.Background())
	if err != nil {
		return err
	}

	// Archive user data
	_, err = tx.Exec(archiveQuery, userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Hard delete
	result, err := tx.Exec(deleteQuery, userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		tx.Rollback()
		return ErrUserNotFound
	}

	return tx.Commit()
}

func (r *UserRepositoryImpl) HardDeletePermanentlyInactiveUsers() error {
	query := `
		DELETE FROM users 
		WHERE 
			status = 'inactive' AND 
			deleted_at < datetime('now', '-365 days')
	`
	_, err := r.db.ExecuteQuery(query)
	return err
}

func (r *UserRepositoryImpl) MarkInactiveUsers() error {
	query := `
		UPDATE users 
		SET 
			status = 'inactive', 
			deleted_at = datetime('now')
		WHERE 
			status = 'active' AND 
			last_activity_at < datetime('now', '-90 days')
	`
	_, err := r.db.ExecuteQuery(query)
	return err
}

func (r *UserRepositoryImpl) LockSecurityViolationUsers() error {
	query := `
		UPDATE users 
		SET 
			status = 'locked', 
			lock_reason = 'Multiple security violations',
			locked_until = datetime('now', '+30 days')
		WHERE 
			id IN (
				SELECT user_id 
				FROM login_attempts 
				WHERE 
					attempts > 10 AND 
					success = false AND 
					last_attempt > datetime('now', '-30 days')
			)
	`
	_, err := r.db.ExecuteQuery(query)
	return err
}

var _ UserRepository = (*UserRepositoryImpl)(nil)
