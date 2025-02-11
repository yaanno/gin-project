package utils

import (
	"regexp"
	"unicode"
)

func IsPasswordComplex(password string) bool {
	// Password complexity rules:
	// - Minimum 12 characters
	// - At least 1 uppercase letter
	// - At least 1 lowercase letter
	// - At least 1 number
	// - At least 1 special character

	if len(password) < 12 {
		return false
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func SanitizePassword(password string) string {
	// Remove any potentially dangerous characters
	reg := regexp.MustCompile(`[^a-zA-Z0-9!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]+`)
	return reg.ReplaceAllString(password, "")
}
