package utils

import (
	"testing"
)

func TestHashPassword(t *testing.T) {

	password := "ValidP@ssw0rd"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Errorf("Unexpected error hashing password: %v", err)
	} else {
		t.Logf("Hashed password successfully: %s", hashedPassword)
	}

	if len(hashedPassword) == 0 {
		t.Error("Hashed password should not be empty")
	}
}

func TestComparePassword(t *testing.T) {

	password := "ValidP@ssw0rd"
	hashedPassword, _ := HashPassword(password)

	tests := []struct {
		name          string
		plainPassword string
		expected      bool
	}{
		{"ValidPassword", password, true},
		{"InvalidPassword", "InvalidP@ss", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComparePassword(hashedPassword, tt.plainPassword)
			if result != tt.expected {
				t.Errorf("ComparePassword() = %v, want %v", result, tt.expected)
			} else {
				t.Logf("Password comparison test passed for: %s", tt.name)
			}
		})
	}
}

func TestHashPasswordError(t *testing.T) {

	shortPassword := "short"
	_, err := HashPassword(shortPassword)
	if err == nil {
		t.Error("Expected error for short password, got none")
	}
}
