package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/user-management-api/pkg/utils"
)

func TestValidatePassword(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		isValid  bool
		minScore int
	}{
		{"Valid Complex Password", "StrongP@ssw0rd2024!", true, 4},
		{"Too Short", "Short1!", false, 0},
		{"No Uppercase", "lowercase1!", false, 2},
		{"No Lowercase", "UPPERCASE1!", false, 2},
		{"No Number", "NoNumberSpecial!", false, 2},
		{"No Special Character", "NoSpecialChar123", false, 3},
		{"Common Password", "password", false, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &utils.PasswordValidatorImpl{}
			result := p.ValidatePassword(tc.password)
			assert.Equal(t, tc.isValid, result.IsValid)
			assert.GreaterOrEqual(t, result.StrengthScore, tc.minScore)
		})
	}
}
