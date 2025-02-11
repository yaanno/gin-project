package repository

import (
	"database/sql"
	"errors"

	"github.com/yourusername/user-management-api/internal/database"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

func CreateUser(user *database.User) error {
	query := `
		INSERT INTO users (username, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id
	`
	err := database.DB.QueryRow(query, user.Username, user.Email, user.Password).Scan(&user.ID)
	return err
}

func FindUserByUsername(username string) (*database.User, error) {
	user := &database.User{}
	query := `
		SELECT id, username, email, password 
		FROM users 
		WHERE username = $1
	`
	err := database.DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func FindUserByID(userID uint) (*database.User, error) {
	user := &database.User{}
	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users 
		WHERE id = $1
	`
	err := database.DB.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func UpdateUser(user *database.User) error {
	query := `
		UPDATE users 
		SET email = $1, password = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := database.DB.Exec(query, user.Email, user.Password, user.ID)
	return err
}

func DeleteUser(userID uint) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := database.DB.Exec(query, userID)
	return err
}

func GetAllUsers() ([]database.User, error) {
	query := `
		SELECT id, username, email, created_at, updated_at 
		FROM users 
		ORDER BY created_at DESC
	`
	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
