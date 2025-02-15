// pkg/token/token.go
package token

import (
	"errors"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType defines different types of tokens
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents the standard claims for our tokens
type Claims struct {
	UserID      uint                   `json:"user_id"`
	Username    string                 `json:"username"`
	TokenType   TokenType              `json:"token_type"`
	Permissions []string               `json:"permissions"`
	DeviceInfo  map[string]interface{} `json:"device_info"`
	jwt.RegisteredClaims
}

// Errors for token management
type (
	TokenExpiredError     struct{ Message string }
	TokenBlacklistedError struct{ Message string }
	InvalidTokenTypeError struct{ Message string }
)

func (e TokenExpiredError) Error() string     { return e.Message }
func (e TokenBlacklistedError) Error() string { return e.Message }
func (e InvalidTokenTypeError) Error() string { return e.Message }

type TokenManager interface {
	ValidateToken(tokenString string, tokenType TokenType) (*Claims, error)
	InvalidateToken(tokenString string) error
	GenerateToken(
		userID uint,
		username string,
		tokenType TokenType,
	) (string, error)
}

// TokenManager handles all token-related operations
type TokenManagerImpl struct {
	secretKey        []byte
	refreshSecretKey []byte
	blacklist        *TokenBlacklist
}

// NewTokenManager creates a new TokenManager
func NewTokenManager(secretKey, refreshSecretKey string) *TokenManagerImpl {
	return &TokenManagerImpl{
		secretKey:        []byte(secretKey),
		refreshSecretKey: []byte(refreshSecretKey),
		blacklist:        NewTokenBlacklist(),
	}
}

// GenerateToken generates a new token with specified type
func (tm *TokenManagerImpl) GenerateToken(
	userID uint,
	username string,
	tokenType TokenType,
) (string, error) {
	var (
		expirationDuration time.Duration
		secretKey          []byte
	)

	switch tokenType {
	case AccessToken:
		expirationDuration = 15 * time.Minute
		secretKey = tm.secretKey
	case RefreshToken:
		expirationDuration = 7 * 24 * time.Hour
		secretKey = tm.refreshSecretKey
	default:
		return "", InvalidTokenTypeError{Message: "Invalid token type"}
	}

	claims := Claims{
		UserID:    userID,
		Username:  username,
		TokenType: tokenType,
		Permissions: []string{
			"read",
			"write",
		},
		DeviceInfo: map[string]interface{}{
			"ip": "127.0.0.1", // TODO: Implement actual device detection
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "user-management-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// ValidateToken validates a token and returns its claims
func (tm *TokenManagerImpl) ValidateToken(
	tokenString string,
	expectedTokenType TokenType,
) (*Claims, error) {
	// Check if token is blacklisted first
	if tm.blacklist.IsBlacklisted(tokenString) {
		return nil, TokenBlacklistedError{Message: "Token is blacklisted"}
	}

	// Determine which secret key to use
	var secretKey []byte
	switch expectedTokenType {
	case AccessToken:
		secretKey = tm.secretKey
	case RefreshToken:
		secretKey = tm.refreshSecretKey
	default:
		return nil, InvalidTokenTypeError{Message: "Invalid token type"}
	}

	// Parse and validate token
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, TokenExpiredError{Message: "Token has expired"}
		}
		return &Claims{}, err
	}

	// Type assert claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return &Claims{}, errors.New("invalid token claims")
	}

	// Validate token type
	if claims.TokenType != expectedTokenType {
		return nil, InvalidTokenTypeError{
			Message: "Token type does not match expected type",
		}
	}

	return claims, nil
}

func (tm *TokenManagerImpl) InvalidateToken(tokenString string) error {
	if tm.blacklist.IsBlacklisted(tokenString) {
		return TokenBlacklistedError{Message: "Token is blacklisted"}
	}

	_, err := tm.ValidateToken(tokenString, "access")
	if err != nil {
		return err
	}

	tm.blacklist.AddToken(tokenString, time.Now().Add(24*time.Hour))
	return nil
}

// Blacklist management
type TokenBlacklist struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

func NewTokenBlacklist() *TokenBlacklist {
	return &TokenBlacklist{
		tokens: make(map[string]time.Time),
	}
}

func (tb *TokenBlacklist) AddToken(token string, expiration time.Time) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.tokens[token] = expiration
}

func (tb *TokenBlacklist) IsBlacklisted(token string) bool {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	expiration, exists := tb.tokens[token]
	return exists && time.Now().Before(expiration)
}

func (tb *TokenBlacklist) Cleanup() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	for token, expiresAt := range tb.tokens {
		if now.After(expiresAt) {
			delete(tb.tokens, token)
		}
	}
}

var _ TokenManager = (*TokenManagerImpl)(nil)
