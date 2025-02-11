package repository_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourusername/user-management-api/internal/database"
	"github.com/yourusername/user-management-api/internal/database/sqlite"
)

func setupSQLiteTestDB(t *testing.T) *sql.DB {
	// Create an in-memory test database
	db, err := sqlite.CreateInMemoryTestDB()
	require.NoError(t, err)

	return db
}

func TestSQLiteUserRepository(t *testing.T) {
	db := setupSQLiteTestDB(t)
	defer db.Close()

	// Implement similar tests as in PostgreSQL repository test
	t.Run("Create User", func(t *testing.T) {
		user := &database.User{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "TestPassword123!",
		}

		// Implement user creation logic for SQLite
		query := `
            INSERT INTO users (username, email, password, created_at, updated_at)
            VALUES (?, ?, ?, datetime('now'), datetime('now'))
        `
		result, err := db.Exec(query, user.Username, user.Email, user.Password)
		require.NoError(t, err)

		// Check that a row was inserted
		rowsAffected, err := result.RowsAffected()
		require.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)
	})

	// Add more tests for other repository methods
}

func setupTestDatabase() *sql.DB {
	// Setup a test database connection
	// In a real-world scenario, use a test database or mock
	db, err := sql.Open("postgres", "connection_string_for_test_db")
	if err != nil {
		panic(err)
	}
	return db
}

// func TestCreateUser(t *testing.T) {
// 	db := setupTestDatabase()
// 	defer db.Close()

// 	user := &database.User{
// 		Username: "testuser",
// 		Email:    "test@example.com",
// 		Password: "TestPassword123!",
// 	}

// 	err := CreateUser(user)
// 	require.NoError(t, err)
// 	assert.NotZero(t, user.ID)
// }

// func TestFindUserByUsername(t *testing.T) {
// 	db := setupTestDatabase()
// 	defer db.Close()

// 	// First, create a test user
// 	user := &database.User{
// 		Username: "finduser",
// 		Email:    "find@example.com",
// 		Password: "TestPassword123!",
// 	}
// 	err := CreateUser(user)
// 	require.NoError(t, err)

// 	// Then try to find the user
// 	foundUser, err := FindUserByUsername("finduser")
// 	require.NoError(t, err)
// 	assert.Equal(t, user.Username, foundUser.Username)
// 	assert.Equal(t, user.Email, foundUser.Email)
// }

// func TestUpdateUser(t *testing.T) {
// 	db := setupTestDatabase()
// 	defer db.Close()

// 	// Create a user to update
// 	user := &database.User{
// 		Username: "updateuser",
// 		Email:    "update@example.com",
// 		Password: "TestPassword123!",
// 	}
// 	err := CreateUser(user)
// 	require.NoError(t, err)

// 	// Update user details
// 	user.Email = "updated@example.com"
// 	err = UpdateUser(user)
// 	require.NoError(t, err)

// 	// Verify update
// 	updatedUser, err := FindUserByID(user.ID)
// 	require.NoError(t, err)
// 	assert.Equal(t, "updated@example.com", updatedUser.Email)
// }

// func TestDeleteUser(t *testing.T) {
// 	db := setupTestDatabase()
// 	defer db.Close()

// 	// Create a user to delete
// 	user := &database.User{
// 		Username: "deleteuser",
// 		Email:    "delete@example.com",
// 		Password: "TestPassword123!",
// 	}
// 	err := CreateUser(user)
// 	require.NoError(t, err)

// 	// Delete the user
// 	err = DeleteUser(user.ID)
// 	require.NoError(t, err)

// 	// Verify deletion
// 	_, err = FindUserByID(user.ID)
// 	assert.Error(t, err)
// 	assert.Equal(t, ErrUserNotFound, err)
// }
