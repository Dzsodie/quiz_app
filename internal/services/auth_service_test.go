package services

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthService_RegisterUser(t *testing.T) {
	s := &AuthService{}

	// Test registering a new user
	err := s.RegisterUser("testuser", "password123")
	assert.NoError(t, err, "expected no error when registering a new user")

	// Test registering the same user again
	err = s.RegisterUser("testuser", "password123")
	assert.Error(t, err, "expected error when registering a duplicate user")
	assert.Equal(t, "user already exists", err.Error(), "unexpected error message")
}

func TestAuthService_AuthenticateUser(t *testing.T) {
	s := &AuthService{}

	// Prepopulate a user
	s.RegisterUser("testuser", "password123")

	// Test authenticating with correct credentials
	err := s.AuthenticateUser("testuser", "password123")
	assert.NoError(t, err, "expected no error when authenticating with valid credentials")

	// Test authenticating with incorrect password
	err = s.AuthenticateUser("testuser", "wrongpassword")
	assert.Error(t, err, "expected error when authenticating with incorrect password")
	assert.Equal(t, "invalid username or password", err.Error(), "unexpected error message")

	// Test authenticating with non-existent username
	err = s.AuthenticateUser("nonexistent", "password123")
	assert.Error(t, err, "expected error when authenticating with non-existent user")
	assert.Equal(t, "invalid username or password", err.Error(), "unexpected error message")
}

func TestAuthService_Concurrency(t *testing.T) {
	s := &AuthService{}

	wg := sync.WaitGroup{}
	numRoutines := 100

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			username := fmt.Sprintf("user%d", i) // Ensure unique usernames
			err := s.RegisterUser(username, "password")
			if err != nil && err.Error() != "user already exists" {
				t.Errorf("unexpected error: %v", err)
			}
		}(i)
	}
	wg.Wait()

	// Verify all unique users are registered
	for i := 0; i < numRoutines; i++ {
		username := fmt.Sprintf("user%d", i)
		err := s.AuthenticateUser(username, "password")
		assert.NoError(t, err, "expected no error when authenticating a registered user")
	}
}
