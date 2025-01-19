package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Dzsodie/quiz_app/internal/database"
	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/Dzsodie/quiz_app/internal/services"
	"github.com/Dzsodie/quiz_app/internal/utils"
	"go.uber.org/zap"
)

type AuthHandler struct {
	AuthService services.IAuthService
	Database    *database.MemoryDB
}

func NewAuthHandler(authService services.IAuthService) *AuthHandler {
	return &AuthHandler{AuthService: authService, Database: database.NewMemoryDB()}
}

// @Summary Register a new user
// @Description Register a user with a username and password
// @Tags User
// @Accept json
// @Produce json
// @Param user body models.User true "User details"
// @Success 201 {object} map[string]string "message"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 409 {object} map[string]string "User already exists"
// @Router /register [post]
func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger().Sugar()

	// Decode request body
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil || user.Username == "" || user.Password == "" {
		logger.Warn("Invalid registration input", zap.Error(err))
		http.Error(w, `{"message":"Invalid input"}`, http.StatusBadRequest)
		return
	}

	// Register user
	if err := h.AuthService.RegisterUser(user.Username, user.Password); err != nil {
		if err.Error() == "user already exists" {
			logger.Warn("User already exists", zap.String("username", user.Username))
			http.Error(w, `{"message":"User already exists"}`, http.StatusConflict)
		} else {
			logger.Error("Error during registration", zap.Error(err))
			http.Error(w, `{"message":"Internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	// Successful registration response
	response := map[string]string{"message": "User registered successfully"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to send response", zap.Error(err))
	}
}

// @Summary Login a user
// @Description Login with a username and password
// @Tags User
// @Accept json
// @Produce json
// @Param user body models.User true "User details"
// @Success 200 {object} map[string]string "message"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Router /login [post]
func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger().Sugar()
	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil || user.Username == "" || user.Password == "" {
		logger.Warn("Invalid input for login", zap.Error(err))
		http.Error(w, `{"message":"Invalid input"}`, http.StatusBadRequest)
		return
	}

	if err := h.AuthService.AuthenticateUser(user.Username, user.Password); err != nil {
		logger.Warn("Authentication failed", zap.String("username", user.Username), zap.Error(err))
		http.Error(w, `{"message":"Invalid username or password"}`, http.StatusUnauthorized)
		return
	}

	if err := utils.ValidateSessionStore(); err != nil {
		logger.Error("Session store not initialized", zap.Error(err))
		http.Error(w, `{"message":"session store - Internal server error"}`, http.StatusInternalServerError)
		return
	}

	sessionToken, err := utils.GenerateSessionToken()
	if err != nil {
		logger.Warn("Failed to generate session token", zap.Error(err))
		http.Error(w, `{"message":" genereate - Internal server error"}`, http.StatusInternalServerError)
		return
	}

	utils.SessionDB[sessionToken] = user.Username

	session, err := utils.SessionStore.Get(r, "quiz-session")
	if err != nil {
		logger.Warn("Failed to retrieve session", zap.Error(err))
		http.Error(w, `{"message":"retrieve - Internal server error"}`, http.StatusInternalServerError)
		return
	}

	session.Values["username"] = user.Username
	session.Values["session_token"] = sessionToken
	userID, err := h.AuthService.GetUserID(user.Username)
	if err != nil {
		logger.Error("Failed to retrieve user ID", zap.Error(err))
		http.Error(w, `{"message":"Internal server error"}`, http.StatusInternalServerError)
		return
	}
	session.Values["userID"] = userID

	if err := session.Save(r, w); err != nil {
		logger.Error("Failed to save session", zap.Error(err))
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}
	logger.Debug("Session saved successfully", zap.Any("session_values", session.Values))

	response := map[string]string{
		"message":       "Login successful",
		"session_token": sessionToken,
		"username":      user.Username,
		"userID":        user.UserID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to send response", zap.Error(err))
	}
}
