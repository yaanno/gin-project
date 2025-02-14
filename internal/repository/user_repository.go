package repository

import (
	"database/sql"
	"errors"

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
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecuteQuery(query, userID)
	if err != nil {
		r.log.Error().Err(err).Msg("Failed to delete user")
		return err
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

var _ UserRepository = (*UserRepositoryImpl)(nil)
