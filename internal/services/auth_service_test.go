package services

import (
	"sync"
	"testing"

	"github.com/Dzsodie/quiz_app/internal/database"
	"github.com/stretchr/testify/assert"
)

func TestAuthServiceRegisterUser(t *testing.T) {
	db := database.NewMemoryDB()
	authService := NewAuthService(db)

	err := authService.RegisterUser("testuser", "Password123!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user, err := db.GetUser("testuser")
	if err != nil || user.Username != "testuser" {
		t.Errorf("user was not correctly registered")
	}
}

func TestAuthServiceAuthenticateUser(t *testing.T) {
	db := database.NewMemoryDB()
	authService := NewAuthService(db)

	// Add a test user
	username := "testuser"
	password := "P@ssw0rd123"
	err := authService.RegisterUser(username, password)
	if err != nil {
		t.Fatalf("unexpected error during user registration: %v", err)
	}

	tests := []struct {
		name          string
		username      string
		password      string
		expectedError bool
	}{
		{
			name:          "Valid credentials",
			username:      "testuser",
			password:      "P@ssw0rd123",
			expectedError: false,
		},
		{
			name:          "Invalid username",
			username:      "invaliduser",
			password:      "P@ssw0rd123",
			expectedError: true,
		},
		{
			name:          "Invalid password",
			username:      "testuser",
			password:      "WrongPassword",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authService.AuthenticateUser(tt.username, tt.password)
			if (err != nil) != tt.expectedError {
				t.Errorf("AuthenticateUser() error = %v, expectedError = %v", err, tt.expectedError)
			}
		})
	}
}

func TestAuthServiceConcurrency(t *testing.T) {
	db := database.NewMemoryDB()
	authService := NewAuthService(db)

	var wg sync.WaitGroup
	numRoutines := 10
	usernameBase := "testuser"
	password := "P@ssw0rd123!"

	// Test concurrent user registration
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(i int) {
			defer wg.Done()
			username := usernameBase + string(rune(i))
			err := authService.RegisterUser(username, password)
			if err != nil && err.Error() != "user already exists" {
				t.Errorf("unexpected error during registration: %v", err)
			}
		}(i)
	}
	wg.Wait()

	// Verify all users were registered
	for i := 0; i < numRoutines; i++ {
		username := usernameBase + string(rune(i))
		user, err := db.GetUser(username)
		if err != nil {
			t.Errorf("user %s was not found: %v", username, err)
		}
		assert.Equal(t, username, user.Username)
	}
}
