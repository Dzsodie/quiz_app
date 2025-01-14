package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/Dzsodie/quiz_app/internal/services"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	AuthService services.IAuthService
}

// NewAuthHandler creates a new instance of AuthHandler with the provided IAuthService implementation.
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
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	err := h.AuthService.RegisterUser(user.Username, user.Password)
	if err != nil {
		if err.Error() == "user already exists" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

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
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	err := h.AuthService.AuthenticateUser(user.Username, user.Password)
	if err != nil {
		if err.Error() == "invalid username or password" {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	session, _ := SessionStore.Get(r, "quiz-session")
	session.Values["username"] = user.Username
	session.Save(r, w)

	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}
