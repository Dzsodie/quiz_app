package services

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// newTestAuthService creates a new instance of AuthService for testing purposes.
func newTestAuthService() *AuthService {
	return &AuthService{}
}

func TestAuthServiceRegisterUser(t *testing.T) {
	s := newTestAuthService()
	err := s.RegisterUser("testuser", "Valid@123")
	assert.NoError(t, err, "expected no error when registering a new user")

	err = s.RegisterUser("testuser", "Valid@123")
	assert.Error(t, err, "expected error when registering a duplicate user")
	assert.Equal(t, "user already exists", err.Error(), "unexpected error message")
}

func TestAuthServiceAuthenticateUser(t *testing.T) {
	s := newTestAuthService()
	uniqueUsername := fmt.Sprintf("testuser_%d", time.Now().UnixNano())

	// Use a valid password with at least one uppercase letter
	err := s.RegisterUser(uniqueUsername, "Password123!")
	if err != nil {
		t.Fatalf("failed to register user for authentication test: %v", err)
	}

	// Test successful authentication
	err = s.AuthenticateUser(uniqueUsername, "Password123!")
	assert.NoError(t, err, "expected no error when authenticating with valid credentials")

	// Test invalid password
	err = s.AuthenticateUser(uniqueUsername, "wrongpassword")
	assert.Error(t, err, "expected error when authenticating with incorrect password")
	assert.Equal(t, "invalid username or password", err.Error(), "unexpected error message")

	// Test non-existent user
	err = s.AuthenticateUser("nonexistent", "Password123!")
	assert.Error(t, err, "expected error when authenticating with non-existent user")
	assert.Equal(t, "invalid username or password", err.Error(), "unexpected error message")
}

func TestAuthServiceConcurrency(t *testing.T) {
	s := newTestAuthService()

	wg := sync.WaitGroup{}
	numRoutines := 50

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			username := fmt.Sprintf("user%d", i)
			err := s.RegisterUser(username, "Valid@123")
			if err != nil && err.Error() != "user already exists" {
				t.Errorf("unexpected error: %v", err)
			}
		}(i)
	}
	wg.Wait()

	for i := 0; i < numRoutines; i++ {
		username := fmt.Sprintf("user%d", i)
		err := s.AuthenticateUser(username, "Valid@123")
		assert.NoError(t, err, "expected no error when authenticating a registered user")
	}
}
