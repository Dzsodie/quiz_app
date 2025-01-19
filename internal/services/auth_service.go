package services

import (
	"errors"
	"fmt"
	"sync"

	"net/http"

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

	// Validate username and password
	if err := utils.ValidateUsername(username); err != nil {
		logger.Warn("Invalid username", zap.String("username", username), zap.Error(err))
		return fmt.Errorf("invalid username: %w", err)
	}

	if err := utils.ValidatePassword(password, "Password must contain at least one uppercase letter"); err != nil {
		logger.Warn("Invalid password", zap.String("username", username), zap.Error(err))
		return fmt.Errorf("invalid password: %w", err)
	}

	authMu.Lock()
	defer authMu.Unlock()

	// Check if user already exists
	_, err := s.DB.GetUser(username)
	if err == nil {
		logger.Warn("User already exists", zap.String("username", username))
		return errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		logger.Error("Error hashing password", zap.String("username", username), zap.Error(err))
		return fmt.Errorf("error hashing password: %w", err)
	}

	// Save user to MemoryDB
	s.DB.AddUser(database.User{
		UserID:     uuid.NewString(),
		Username:   username,
		Password:   hashedPassword,
		Progress:   []int{},
		Score:      0,
		QuizTaken:  0,
		Percentage: 0,
	})

	logger.Info("User registered successfully", zap.String("username", username))
	return nil
}

func (s *AuthService) AuthenticateUser(username, password string) error {
	logger := utils.GetLogger().Sugar()
	authMu.Lock()
	defer authMu.Unlock()

	// Query user from MemoryDB
	user, err := s.DB.GetUser(username)
	if err != nil {
		logger.Warn("Authentication failed: user does not exist", zap.String("username", username))
		return errors.New("invalid username or password")
	}

	// Compare the stored hashed password with the provided password
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

func (s *AuthService) GetSession(r *http.Request) (*sessions.Session, error) {
	logger := utils.GetLogger().Sugar()

	if err := utils.ValidateSessionStore(); err != nil {
		logger.Error("Session store not initialized", zap.Error(err))
		return nil, errors.New("session store is not initialized")
	}

	session, err := utils.SessionStore.Get(r, "quiz-session")
	if err != nil {
		logger.Warn("Failed to get session", zap.Error(err))
		return nil, errors.New("failed to get session")
	}

	logger.Debug("Session retrieved successfully", zap.Any("session_values", session.Values))
	return session, nil
}

var _ IAuthService = &AuthService{}
