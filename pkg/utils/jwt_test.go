package utils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTTokenGeneration(t *testing.T) {
	userID := uint(1)
	username := "testuser"

	// Test Access Token
	accessToken, err := GenerateAccessToken(context.Background(), userID, username)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)

	// Validate Access Token
	claims, err := ValidateToken(context.Background(), accessToken, "access")
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, "access", claims.TokenType)

	// Test Refresh Token
	refreshToken, err := GenerateRefreshToken(context.Background(), userID, username)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshToken)

	// Validate Refresh Token
	refreshClaims, err := ValidateToken(context.Background(), refreshToken, "refresh")
	require.NoError(t, err)
	assert.Equal(t, userID, refreshClaims.UserID)
	assert.Equal(t, username, refreshClaims.Username)
	assert.Equal(t, "refresh", refreshClaims.TokenType)
}

func TestTokenExpiration(t *testing.T) {
	userID := uint(1)
	username := "testuser"

	// Test Access Token Expiration
	accessToken, err := GenerateAccessToken(context.Background(), userID, username)
	require.NoError(t, err)

	claims, err := ValidateToken(context.Background(), accessToken, "access")
	require.NoError(t, err)

	// Check expiration is within 24 hours
	expirationTime := claims.ExpiresAt.Time
	assert.True(t, expirationTime.After(time.Now()))
	assert.True(t, expirationTime.Before(time.Now().Add(25*time.Hour)))
}
