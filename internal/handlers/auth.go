package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/Dzsodie/quiz_app/internal/services"
	"go.uber.org/zap"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	AuthService services.IAuthService
	Logger      *zap.SugaredLogger
}

// NewAuthHandler creates a new instance of AuthHandler with the provided IAuthService implementation.
func NewAuthHandler(authService services.IAuthService, logger *zap.SugaredLogger) *AuthHandler {
	return &AuthHandler{AuthService: authService, Logger: logger}
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
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.Logger.Warn("Invalid input for registration", zap.Error(err))
		http.Error(w, `{"message":"Invalid input"}`, http.StatusBadRequest)
		return
	}

	if err := h.AuthService.RegisterUser(user.Username, user.Password); err != nil {
		if err.Error() == "user already exists" {
			h.Logger.Warn("User already exists", zap.String("username", user.Username))
			w.WriteHeader(http.StatusConflict)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"message": "user already exists"})
		} else {
			h.Logger.Error("Internal server error during registration", zap.Error(err))
			http.Error(w, `{"message":"Internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	h.Logger.Info("User registered successfully", zap.String("username", user.Username))
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
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
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil || user.Username == "" || user.Password == "" {
		h.Logger.Warn("Invalid input for login", zap.Error(err))
		http.Error(w, `{"message":"Invalid input"}`, http.StatusBadRequest)
		return
	}

	if err := h.AuthService.AuthenticateUser(user.Username, user.Password); err != nil {
		if err.Error() == "invalid username or password" {
			h.Logger.Warn("Invalid username or password", zap.String("username", user.Username))
			http.Error(w, `{"message":"invalid username or password"}`, http.StatusUnauthorized)
		} else {
			h.Logger.Error("Internal server error during login", zap.Error(err))
			http.Error(w, `{"message":"Internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	session, _ := SessionStore.Get(r, "quiz-session")
	session.Values["username"] = user.Username
	session.Save(r, w)
	h.Logger.Info("User logged in successfully", zap.String("username", user.Username))
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}
