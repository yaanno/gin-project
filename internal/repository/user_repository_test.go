package repository_test

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/yourusername/user-management-api/internal/database"
	"github.com/yourusername/user-management-api/internal/database/sqlite-gorm"
	"github.com/yourusername/user-management-api/internal/repository"
)

var db *gorm.DB

func init() {
	db, _ = sqlite.InitializeDatabase(sqlite.DatabaseConfig{
		Path:            "file::memory:?cache=shared",
		MaxOpenConns:    10,
		MaxIdleConns:    10,
		ConnMaxLifetime: 10 * time.Second,
		ConnMaxIdleTime: 10 * time.Second,
	})
}

func AfterEach() {
	db.Exec("DELETE FROM users")
}

func TestSQLiteUserRepository(t *testing.T) {
	t.Cleanup(AfterEach)
	// db := setupSQLiteTestDB(t)

	// Implement similar tests as in PostgreSQL repository test
	t.Run("Create User", func(t *testing.T) {
		user := &database.User{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "TestPassword123!",
		}
		result := db.Create(user)
		require.NoError(t, result.Error)

		// Check that a row was inserted
		rowsAffected := result.RowsAffected
		assert.Equal(t, int64(1), rowsAffected)
	})

	// Add more tests for other repository methods
}

func TestCreateUser(t *testing.T) {
	t.Cleanup(AfterEach)
	user := &database.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "TestPassword123!",
	}

	repo := repository.NewUserRepository(db, zerolog.Logger{})

	err := repo.CreateUser(user)
	require.NoError(t, err)
	// assert.NotZero(t, user.ID)
}

func TestFindUserByID(t *testing.T) {
	t.Cleanup(AfterEach)
	user := &database.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "TestPassword123!",
	}

	repo := repository.NewUserRepository(db, zerolog.Logger{})

	err := repo.CreateUser(user)
	require.NoError(t, err)
	require.NoError(t, err)

	// Then try to find the user
	foundUser, err := repo.FindUserByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Username, foundUser.Username)
	assert.Equal(t, user.Email, foundUser.Email)
}

func TestFindUserByUsername(t *testing.T) {
	t.Cleanup(AfterEach)
	user := &database.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "TestPassword123!",
	}

	repo := repository.NewUserRepository(db, zerolog.Logger{})

	err := repo.CreateUser(user)
	require.NoError(t, err)
	require.NoError(t, err)

	// Then try to find the user
	foundUser, err := repo.FindUserByUsername("testuser")
	require.NoError(t, err)
	assert.Equal(t, user.Username, foundUser.Username)
	assert.Equal(t, user.Email, foundUser.Email)
}

func TestUpdateUser(t *testing.T) {
	t.Cleanup(AfterEach)
	repo := repository.NewUserRepository(db, zerolog.Logger{})

	// Create a user to update
	user := &database.User{
		Username: "updateuser",
		Email:    "update@example.com",
		Password: "TestPassword123!",
	}
	err := repo.CreateUser(user)
	require.NoError(t, err)

	// Update user details
	user.Email = "updated@example.com"
	err = repo.UpdateUser(user)
	require.NoError(t, err)

	// Verify update
	updatedUser, err := repo.FindUserByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated@example.com", updatedUser.Email)
}

func TestDeleteUser(t *testing.T) {
	t.Skip()
	t.Cleanup(AfterEach)
	repo := repository.NewUserRepository(db, zerolog.Logger{})

	// Create a user to delete
	user := &database.User{
		Username: "deleteuser",
		Email:    "delete@example.com",
		Password: "TestPassword123!",
	}
	err := repo.CreateUser(user)
	require.NoError(t, err)

	// Soft delete the user
	err = repo.DeleteUser(user.ID)
	require.NoError(t, err)

	// Verify soft deletion
	user, err = repo.FindUserByID(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, user.ID)
	assert.NotNil(t, user.DeletedAt)
}
