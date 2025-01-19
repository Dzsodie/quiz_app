package services

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Dzsodie/quiz_app/internal/database"
	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/Dzsodie/quiz_app/internal/utils"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
)

var (
	authMu sync.Mutex
	users  = make(map[string]models.User)
)

type AuthService struct {
	DB *database.MemoryDB
}

func NewAuthService(db *database.MemoryDB) *AuthService {
	return &AuthService{DB: db}
}

func (s *AuthService) RegisterUser(username, password string) error {
	logger := utils.GetLogger().Sugar()

	if err := utils.ValidateUsername(username); err != nil {
		logger.Warn("User registration failed: invalid username", zap.String("username", username), zap.Error(err))
		return fmt.Errorf("invalid username: %w", err)
	}

	if err := utils.ValidatePassword(password, "Password must contain at least one uppercase letter"); err != nil {
		logger.Warn("User registration failed: invalid password", zap.String("username", username), zap.Error(err))
		return fmt.Errorf("invalid password: %w", err)
	}

	authMu.Lock()
	defer authMu.Unlock()

	if _, exists := users[username]; exists {
		logger.Warn("User registration failed: user already exists", zap.String("username", username))
		return errors.New("user already exists")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		logger.Error("User registration failed: error hashing password", zap.String("username", username), zap.Error(err))
		return fmt.Errorf("error hashing password: %w", err)
	}

	userID := uuid.NewString()
	users[username] = models.User{
		UserID:   userID,
		Username: username,
		Password: hashedPassword,
		Progress: []int{},
		Score:    0,
	}

	logger.Info("User registered successfully", zap.String("username", username), zap.String("userID", userID))
	return nil
}

func (s *AuthService) AuthenticateUser(username, password string) error {
	logger := utils.GetLogger().Sugar()
	authMu.Lock()
	defer authMu.Unlock()

	user, exists := users[username]
	if !exists {
		logger.Warn("Authentication failed: user does not exist", zap.String("username", username))
		return errors.New("invalid username or password")
	}

	if !utils.ComparePassword(user.Password, password) {
		logger.Warn("Authentication failed: invalid password", zap.String("username", username))
		return errors.New("invalid username or password")
	}

	logger.Info("User authenticated successfully", zap.String("username", username))
	return nil
}

func (s *AuthService) GetUserID(username string) (string, error) {
	authMu.Lock()
	defer authMu.Unlock()

	user, exists := users[username]
	if !exists {
		return "", errors.New("user not found")
	}
	return user.UserID, nil
}

func (a *AuthService) GetSession() (*sessions.Session, error) {

	session, err := utils.SessionStore.Get(nil, "session-name")
	if err != nil {
		return nil, errors.New("failed to get session")
	}
	return session, nil
}

var _ IAuthService = &AuthService{}
