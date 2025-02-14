package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourusername/user-management-api/internal/database/sqlite"
	"github.com/yourusername/user-management-api/internal/repository"
	"github.com/yourusername/user-management-api/internal/services"
	"github.com/yourusername/user-management-api/pkg/token"
)

var db, _ = sqlite.CreateInMemoryTestDB()
var repo = repository.NewUserRepository(db, zerolog.Logger{})
var loginAttemptRepo = repository.NewLoginAttemptRepository(db, zerolog.Logger{})
var tokenManager = token.NewTokenManager("secret_key", "refresh_secret_key")
var authService = services.NewAuthService(tokenManager, repo, loginAttemptRepo, zerolog.Logger{})
var authHandler = NewAuthHandler(authService, zerolog.Logger{})

func setupTestRouter() *gin.Engine {
	router := gin.Default()
	router.POST("/auth/register", authHandler.RegisterUser)
	router.POST("/auth/login", authHandler.LoginUser)
	return router
}

func TestRegisterUser(t *testing.T) {
	router := setupTestRouter()

	registrationPayload := map[string]string{
		"username": "newuser",
		"email":    "newuser@example.com",
		"password": "StrongP@ssw0rd2024!",
	}
	jsonPayload, _ := json.Marshal(registrationPayload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify user was created in the database
	user, err := repo.FindUserByUsername("newuser")
	require.NoError(t, err)
	assert.Equal(t, "newuser@example.com", user.Email)
}

func TestRegisterUserWithoutEmail(t *testing.T) {
	router := setupTestRouter()

	registrationPayload := map[string]string{
		"username": "newuser",
		"password": "StrongP@ssw0rd2024!",
	}
	jsonPayload, _ := json.Marshal(registrationPayload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterUserWithoutPassword(t *testing.T) {
	router := setupTestRouter()

	registrationPayload := map[string]string{
		"username": "newuser",
		"email":    "newuser@example.com",
	}
	jsonPayload, _ := json.Marshal(registrationPayload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterUserWithoutUsername(t *testing.T) {
	router := setupTestRouter()
	registrationPayload := map[string]string{
		"password": "StrongP@ssw0rd2024!",
		"email":    "newuser@example.com",
	}
	jsonPayload, _ := json.Marshal(registrationPayload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
func TestRegisterUserWithWeakPassword(t *testing.T) {
	router := setupTestRouter()

	registrationPayload := map[string]string{
		"password": "weakpassword",
		"email":    "newuser@example.com",
		"username": "newuser",
	}
	jsonPayload, _ := json.Marshal(registrationPayload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginUser(t *testing.T) {
	// First, register a user
	router := setupTestRouter()
	registrationPayload := map[string]string{
		"username": "loginuser",
		"email":    "loginuser@example.com",
		"password": "StrongP@ssw0rd2024!",
	}
	jsonRegistration, _ := json.Marshal(registrationPayload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonRegistration))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Then test login
	loginPayload := map[string]string{
		"username": "loginuser",
		"password": "StrongP@ssw0rd2024!",
	}
	jsonLogin, _ := json.Marshal(loginPayload)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonLogin))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify token response
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "access_token")
	assert.Contains(t, response, "refresh_token")
}
