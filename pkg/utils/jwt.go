package utils

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(ctx context.Context, userID uint, username string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	secretKey := []byte(os.Getenv("JWT_SECRET"))

	claims := JWTClaims{
		UserID:    userID,
		Username:  username,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "user-management-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func GenerateRefreshToken(ctx context.Context, userID uint, username string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	secretKey := []byte(os.Getenv("JWT_REFRESH_SECRET"))

	claims := JWTClaims{
		UserID:    userID,
		Username:  username,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "user-management-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func ValidateToken(ctx context.Context, tokenString string, tokenType string) (*JWTClaims, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	secretKey := []byte(os.Getenv("JWT_SECRET"))
	if tokenType == "refresh" {
		secretKey = []byte(os.Getenv("JWT_REFRESH_SECRET"))
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		if claims.TokenType != tokenType {
			return nil, fmt.Errorf("invalid token type")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
