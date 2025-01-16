package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/Dzsodie/quiz_app/internal/services"
	"github.com/Dzsodie/quiz_app/internal/utils"
	"go.uber.org/zap"
)

type AuthHandler struct {
	AuthService services.IAuthService
}

func NewAuthHandler(authService services.IAuthService) *AuthHandler {
	return &AuthHandler{AuthService: authService}
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
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil || user.Username == "" || user.Password == "" {
		logger.Warn("Invalid input for registration", zap.Error(err))
		http.Error(w, `{"message":"Invalid input"}`, http.StatusBadRequest)
		return
	}

	if err := h.AuthService.RegisterUser(user.Username, user.Password); err != nil {
		if err.Error() == "user already exists" {
			logger.Warn("User already exists", zap.String("username", user.Username))
			http.Error(w, `{"message":"user already exists"}`, http.StatusConflict)
		} else {
			logger.Error("Internal server error during registration", zap.Error(err))
			http.Error(w, `{"message":"Internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	logger.Info("User registered successfully", zap.String("username", user.Username))
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
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
		if err.Error() == "invalid username or password" {
			logger.Warn("Invalid username or password", zap.String("username", user.Username))
			http.Error(w, `{"message":"invalid username or password"}`, http.StatusUnauthorized)
		} else {
			logger.Error("Internal server error during login", zap.Error(err))
			http.Error(w, `{"message":"Internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	session, err := SessionStore.Get(r, "quiz-session")
	if err != nil {
		logger.Warn("Failed to retrieve session", zap.Error(err))
		http.Error(w, `{"message":"Internal server error"}`, http.StatusInternalServerError)
		return
	}
	session.Values["username"] = user.Username
	if err := session.Save(r, w); err != nil {
		logger.Warn("Failed to save session", zap.Error(err))
		http.Error(w, `{"message":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	logger.Info("User logged in successfully", zap.String("username", user.Username))
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}
