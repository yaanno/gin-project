package utils

import (
	"regexp"
	"unicode"
)

var (
	// Predefined password validation rules
	minPasswordLength = 12
	maxPasswordLength = 64
)

// PasswordValidationResult provides detailed feedback about password strength
type PasswordValidationResult struct {
	IsValid       bool
	Errors        []string
	StrengthScore int
}

// ValidatePassword provides comprehensive password validation
func ValidatePassword(password string) PasswordValidationResult {
	result := PasswordValidationResult{
		IsValid:       true,
		Errors:        []string{},
		StrengthScore: 0,
	}

	// Length check
	if len(password) < minPasswordLength {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must be at least 12 characters long")
	}
	if len(password) > maxPasswordLength {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must not exceed 64 characters")
	}

	// Character type checks
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
			result.StrengthScore++
		case unicode.IsLower(char):
			hasLower = true
			result.StrengthScore++
		case unicode.IsNumber(char):
			hasNumber = true
			result.StrengthScore++
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
			result.StrengthScore++
		}
	}

	// Check character diversity
	if !hasUpper {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must contain at least one uppercase letter")
	}
	if !hasLower {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must contain at least one lowercase letter")
	}
	if !hasNumber {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must contain at least one number")
	}
	if !hasSpecial {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must contain at least one special character")
	}

	// Common password check (optional, can expand this list)
	commonPasswords := []string{"password", "123456", "qwerty"}
	for _, common := range commonPasswords {
		if password == common {
			result.IsValid = false
			result.Errors = append(result.Errors, "Password is too common")
			break
		}
	}

	return result
}

// SanitizePassword removes potentially dangerous characters
func SanitizePassword(password string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]+`)
	return reg.ReplaceAllString(password, "")
}

// IsPasswordComplex is a quick check for password complexity
func IsPasswordComplex(password string) bool {
	result := ValidatePassword(password)
	return result.IsValid
}
