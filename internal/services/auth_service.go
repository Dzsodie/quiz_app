package services

import (
	"errors"
	"sync"

	"github.com/Dzsodie/quiz_app/internal/models"
)

var (
	users  = make(map[string]models.User)
	authMu sync.Mutex
)

// RegisterUser registers a new user.
func RegisterUser(username, password string) error {
	authMu.Lock()
	defer authMu.Unlock()

	if _, exists := users[username]; exists {
		return errors.New("user already exists")
	}

	users[username] = models.User{
		Username: username,
		Password: password,
	}
	return nil
}

// AuthenticateUser validates the username and password.
func AuthenticateUser(username, password string) error {
	authMu.Lock()
	defer authMu.Unlock()

	user, exists := users[username]
	if !exists || user.Password != password {
		return errors.New("invalid username or password")
	}

	return nil
}
