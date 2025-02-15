package authentication_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yourusername/user-management-api/internal/database"
	"github.com/yourusername/user-management-api/pkg/authentication"
	"github.com/yourusername/user-management-api/pkg/token"
)

// Mock Repositories
type MockUserRepository struct {
	mock.Mock
}

// DeleteUser implements repository.UserRepository.
func (m *MockUserRepository) DeleteUser(userID uint) error {
	panic("unimplemented")
}

// FindUserByID implements repository.UserRepository.
func (m *MockUserRepository) FindUserByID(userID uint) (*database.User, error) {
	panic("unimplemented")
}

// GetAllUsers implements repository.UserRepository.
func (m *MockUserRepository) GetAllUsers() ([]database.User, error) {
	panic("unimplemented")
}

// HardDeletePermanentlyInactiveUsers implements repository.UserRepository.
func (m *MockUserRepository) HardDeletePermanentlyInactiveUsers() error {
	panic("unimplemented")
}

// HardDeleteUser implements repository.UserRepository.
func (m *MockUserRepository) HardDeleteUser(userID uint) error {
	panic("unimplemented")
}

// LockSecurityViolationUsers implements repository.UserRepository.
func (m *MockUserRepository) LockSecurityViolationUsers() error {
	panic("unimplemented")
}

// MarkInactiveUsers implements repository.UserRepository.
func (m *MockUserRepository) MarkInactiveUsers() error {
	panic("unimplemented")
}

// MarkUserInactive implements repository.UserRepository.
func (m *MockUserRepository) MarkUserInactive(userID uint) error {
	panic("unimplemented")
}

// UnlockUser implements repository.UserRepository.
func (m *MockUserRepository) UnlockUser(userID uint) error {
	panic("unimplemented")
}

// UpdateUser implements repository.UserRepository.
func (m *MockUserRepository) UpdateUser(user *database.User) error {
	panic("unimplemented")
}

func (m *MockUserRepository) CreateUser(user *database.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindUserByUsername(username string) (*database.User, error) {
	args := m.Called(username)
	return args.Get(0).(*database.User), args.Error(1)
}

func (m *MockUserRepository) LockUser(userID uint, reason string, duration time.Duration) error {
	args := m.Called(userID, reason, duration)
	return args.Error(0)
}

type MockLoginAttemptRepository struct {
	mock.Mock
}

func (m *MockLoginAttemptRepository) GetLoginAttempts(username, ipAddress string) (int, time.Time, error) {
	args := m.Called(username, ipAddress)
	return args.Int(0), args.Get(1).(time.Time), args.Error(2)
}

func (m *MockLoginAttemptRepository) IncrementLoginAttempts(username, ipAddress string, success bool) error {
	args := m.Called(username, ipAddress, success)
	return args.Error(0)
}

func (m *MockLoginAttemptRepository) ResetLoginAttempts(username, ipAddress string) error {
	args := m.Called(username, ipAddress)
	return args.Error(0)
}

// MockTokenManager implements token.TokenManager interface
type MockTokenManager struct {
	mock.Mock
}

func (m *MockTokenManager) GenerateToken(
	userID uint,
	username string,
	tokenType token.TokenType,
) (string, error) {
	args := m.Called(userID, username, tokenType)
	return args.String(0), args.Error(1)
}

func (m *MockTokenManager) ValidateToken(
	tokenString string,
	tokenType token.TokenType,
) (*token.Claims, error) {
	args := m.Called(tokenString, tokenType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*token.Claims), args.Error(1)
}

func (m *MockTokenManager) InvalidateToken(tokenString string) error {
	args := m.Called(tokenString)
	return args.Error(0)
}

// Ensure MockTokenManager implements the TokenManager interface
var _ token.TokenManager = (*MockTokenManager)(nil)

// Test Authentication Manager
// func TestValidateUserAuthentication(t *testing.T) {
// 	testCases := []struct {
// 		name           string
// 		setupMocks     func(*MockUserRepository, *MockLoginAttemptRepository, *MockTokenManager)
// 		username       string
// 		password       string
// 		ipAddress      string
// 		expectedError  bool
// 		expectedErrMsg string
// 	}{
// 		{
// 			name: "Successful Authentication",
// 			setupMocks: func(
// 				userRepo *MockUserRepository,
// 				loginAttemptRepo *MockLoginAttemptRepository,
// 				tokenManager *MockTokenManager,
// 			) {
// 				user := &database.User{
// 					ID:       1,
// 					Username: "validuser",
// 					Status:   database.UserStatusActive,
// 					Password: "hashedpassword", // Assume this is a valid bcrypt hash
// 				}
// 				userRepo.On("FindUserByUsername", "validuser").Return(user, nil)
// 				userRepo.On("CheckPasswordHash", "correctpassword").Return(true)
// 				// Mock login attempt tracking
// 				loginAttemptRepo.On("GetLoginAttempts", "validuser", "127.0.0.1").Return(0, time.Now(), nil)
// 				loginAttemptRepo.On("IncrementLoginAttempts", "validuser", "127.0.0.1", false).Return(nil)
// 				loginAttemptRepo.On("ResetLoginAttempts", "validuser", "127.0.0.1").Return(nil)

// 				// Add token generation mock if needed
// 				tokenManager.On("GenerateToken",
// 					user.ID,
// 					user.Username,
// 					token.AccessToken,
// 				).Return("mock_access_token", nil)
// 			},
// 			username:      "validuser",
// 			password:      "correctpassword",
// 			ipAddress:     "127.0.0.1",
// 			expectedError: false,
// 		},
// 		{
// 			name: "Invalid Password",
// 			setupMocks: func(userRepo *MockUserRepository, loginAttemptRepo *MockLoginAttemptRepository, tokenManager *MockTokenManager) {
// 				user := &database.User{
// 					ID:       1,
// 					Username: "validuser",
// 					Status:   database.UserStatusActive,
// 					Password: "hashedpassword",
// 				}
// 				userRepo.On("FindUserByUsername", "validuser").Return(user, nil)
// 				userRepo.On("CheckPasswordHash", "wrongpassword").Return(false)
// 				// Mock login attempt tracking for failed login
// 				loginAttemptRepo.On("GetLoginAttempts", "validuser", "127.0.0.1").Return(0, time.Now(), nil)
// 				loginAttemptRepo.On("IncrementLoginAttempts", "validuser", "127.0.0.1", false).Return(nil)
// 			},
// 			username:       "validuser",
// 			password:       "wrongpassword",
// 			ipAddress:      "127.0.0.1",
// 			expectedError:  true,
// 			expectedErrMsg: "invalid credentials",
// 		},
// 		{
// 			name: "Locked Account",
// 			setupMocks: func(userRepo *MockUserRepository, loginAttemptRepo *MockLoginAttemptRepository, tokenManager *MockTokenManager) {
// 				user := &database.User{
// 					ID:          1,
// 					Username:    "lockeduser",
// 					Status:      database.UserStatusLocked,
// 					LockedUntil: time.Now().Add(1 * time.Hour),
// 					LockReason:  "Too many login attempts",
// 				}
// 				userRepo.On("FindUserByUsername", "lockeduser").Return(user, nil)
// 			},
// 			username:       "lockeduser",
// 			password:       "anypassword",
// 			ipAddress:      "127.0.0.1",
// 			expectedError:  true,
// 			expectedErrMsg: "account locked until",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// Create mock repositories
// 			userRepo := new(MockUserRepository)
// 			loginAttemptRepo := new(MockLoginAttemptRepository)
// 			tokenManager := new(MockTokenManager)
// 			logger := zerolog.Nop()

// 			// Setup mocks
// 			tc.setupMocks(userRepo, loginAttemptRepo, tokenManager)

// 			// Create Authentication Manager
// 			authManager := authentication.NewAuthenticationManager(
// 				userRepo,
// 				tokenManager,
// 				loginAttemptRepo,
// 				logger,
// 			)

// 			// Perform authentication
// 			user, err := authManager.ValidateUserAuthentication(
// 				context.Background(),
// 				tc.username,
// 				tc.password,
// 				tc.ipAddress,
// 			)

// 			fmt.Println(user, err)

// 			if tc.expectedError {
// 				assert.Error(t, err)
// 				if tc.expectedErrMsg != "" {
// 					assert.Contains(t, err.Error(), tc.expectedErrMsg)
// 				}
// 				assert.Nil(t, user)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, user)
// 			}

// 			// Verify mock expectations
// 			tokenManager.AssertExpectations(t)
// 			userRepo.AssertExpectations(t)
// 			loginAttemptRepo.AssertExpectations(t)
// 		})
// 	}
// }

func TestCalculateLockDelay(t *testing.T) {
	authManager := authentication.NewAuthenticationManager(
		nil,           // userRepo
		nil,           // tokenManager
		nil,           // loginAttemptRepo
		zerolog.Nop(), // logger
	)

	testCases := []struct {
		attempts int
		minDelay time.Duration
		maxDelay time.Duration
	}{
		{1, 1 * time.Second, 2 * time.Second},
		{2, 2 * time.Second, 4 * time.Second},
		{3, 4 * time.Second, 8 * time.Second},
		{4, 8 * time.Second, 16 * time.Second},
		{5, 16 * time.Second, 32 * time.Second},
		{6, 32 * time.Second, 1 * time.Minute},
		{7, 1 * time.Minute, 2 * time.Minute},
		{8, 2 * time.Minute, 4 * time.Minute},
		{9, 4 * time.Minute, 1 * time.Hour},
		{10, 1 * time.Hour, 1 * time.Hour},
		{11, 1 * time.Hour, 1 * time.Hour}, // Test beyond 10 attempts
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Attempts_%d", tc.attempts), func(t *testing.T) {
			delay := authManager.CalculateLockDelay(tc.attempts)

			// Ensure delay is within expected range
			assert.GreaterOrEqual(t, delay, tc.minDelay)
			assert.LessOrEqual(t, delay, tc.maxDelay)
		})
	}
}

func TestCheckLoginAttempts(t *testing.T) {
	testCases := []struct {
		name           string
		attempts       int
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:           "Within Attempt Limit",
			attempts:       4,
			expectedError:  false,
			expectedErrMsg: "",
		},
		{
			name:           "Exceeded Attempt Limit",
			attempts:       5,
			expectedError:  true,
			expectedErrMsg: "too many login attempts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := new(MockUserRepository)
			loginAttemptRepo := new(MockLoginAttemptRepository)
			tokenManager := new(MockTokenManager)
			logger := zerolog.Nop()

			// Setup mocks
			loginAttemptRepo.On("GetLoginAttempts", "testuser", "127.0.0.1").
				Return(tc.attempts, time.Now(), nil)

			if tc.expectedError {
				userRepo.On("LockUser",
					mock.AnythingOfType("uint"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("time.Duration"),
				).Return(nil)
			}

			authManager := authentication.NewAuthenticationManager(
				userRepo,
				tokenManager,
				loginAttemptRepo,
				logger,
			)

			err := authManager.CheckLoginAttempts("testuser", 1, "127.0.0.1")

			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}

			userRepo.AssertExpectations(t)
			loginAttemptRepo.AssertExpectations(t)
		})
	}
}

func TestValidateToken(t *testing.T) {
	testCases := []struct {
		name          string
		tokenString   string
		tokenType     token.TokenType
		mockValidate  func(*MockTokenManager)
		expectedError bool
	}{
		{
			name:        "Valid Access Token",
			tokenString: "valid_access_token",
			tokenType:   token.AccessToken,
			mockValidate: func(m *MockTokenManager) {
				m.On("ValidateToken", "valid_access_token", token.AccessToken).
					Return(&token.Claims{
						UserID:    1,
						Username:  "testuser",
						TokenType: token.AccessToken,
					}, nil)
			},
			expectedError: false,
		},
		{
			name:        "Invalid Token",
			tokenString: "invalid_token",
			tokenType:   token.AccessToken,
			mockValidate: func(m *MockTokenManager) {
				m.On("ValidateToken", "invalid_token", token.AccessToken).
					Return(nil, errors.New("invalid token"))
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenManager := new(MockTokenManager)
			logger := zerolog.Nop()

			// Setup mocks
			tc.mockValidate(tokenManager)

			authManager := authentication.NewAuthenticationManager(
				nil,
				tokenManager,
				nil,
				logger,
			)

			claims, err := authManager.ValidateToken(tc.tokenString, tc.tokenType)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
			}

			tokenManager.AssertExpectations(t)
		})
	}
}
