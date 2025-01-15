package services

import (
	"errors"
	"sync"

	"github.com/Dzsodie/quiz_app/internal/models"
	"go.uber.org/zap"
)

var (
	users  = make(map[string]models.User)
	authMu sync.Mutex
)

type AuthService struct {
	Logger *zap.Logger
}

func NewAuthService(logger *zap.Logger) *AuthService {
	return &AuthService{Logger: logger}
}

// RegisterUser registers a new user.
func (s *AuthService) RegisterUser(username, password string) error {
	authMu.Lock()
	defer authMu.Unlock()

	if _, exists := users[username]; exists {
		s.Logger.Warn("User registration failed: user already exists", zap.String("username", username))
		return errors.New("user already exists")
	}

	users[username] = models.User{
		Username: username,
		Password: password,
	}
	s.Logger.Info("User registered successfully", zap.String("username", username))
	return nil
}

// AuthenticateUser validates the username and password.
func (s *AuthService) AuthenticateUser(username, password string) error {
	authMu.Lock()
	defer authMu.Unlock()

	user, exists := users[username]
	if !exists {
		s.Logger.Warn("Authentication failed: user does not exist", zap.String("username", username))
		return errors.New("invalid username or password")
	}
	if user.Password != password {
		s.Logger.Warn("Authentication failed: invalid password", zap.String("username", username))
		return errors.New("invalid username or password")
	}
	s.Logger.Info("User authenticated successfully", zap.String("username", username))
	return nil
}

var _ IAuthService = &AuthService{}
