package services

import (
	"errors"
	"sync"

	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/Dzsodie/quiz_app/internal/utils"
	"go.uber.org/zap"
)

var (
	authMu sync.Mutex
	users  = make(map[string]models.User)
)

type AuthService struct {
	Logger *zap.Logger
}

func NewAuthService(logger *zap.Logger) *AuthService {
	return &AuthService{Logger: logger}
}

// RegisterUser registers a new user.
func (s *AuthService) RegisterUser(username, password string) error {
	if s.Logger == nil {
		panic("AuthService logger is not set")
	}

	// Validate username
	if err := utils.ValidateUsername(username); err != nil {
		s.Logger.Warn("User registration failed: invalid username", zap.Error(err))
		return err
	}

	// Validate password
	if err := utils.ValidatePassword(password, ""); err != nil {
		s.Logger.Warn("User registration failed: invalid password", zap.Error(err))
		return err
	}

	authMu.Lock()
	defer authMu.Unlock()

	if _, exists := users[username]; exists {
		s.Logger.Warn("User registration failed: user already exists", zap.String("username", username))
		return errors.New("user already exists")
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		s.Logger.Error("User registration failed: error hashing password", zap.Error(err))
		return err
	}

	users[username] = models.User{
		Username: username,
		Password: hashedPassword,
	}
	s.Logger.Info("User registered successfully", zap.String("username", username))
	return nil
}

// AuthenticateUser validates the username and password.
func (s *AuthService) AuthenticateUser(username, password string) error {
	if s.Logger == nil {
		panic("AuthService logger is not set")
	}

	authMu.Lock()
	defer authMu.Unlock()

	user, exists := users[username]
	if !exists {
		s.Logger.Warn("Authentication failed: user does not exist", zap.String("username", username))
		return errors.New("invalid username or password")
	}

	// Compare hashed password
	if !utils.ComparePassword(user.Password, password) {
		s.Logger.Warn("Authentication failed: invalid password", zap.String("username", username))
		return errors.New("invalid username or password")
	}

	s.Logger.Info("User authenticated successfully", zap.String("username", username))
	return nil
}

var _ IAuthService = &AuthService{}
