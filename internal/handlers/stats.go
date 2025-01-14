package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Dzsodie/quiz_app/internal/services"
)

// @Summary Get user statistics
// @Description Retrieve user performance statistics
// @Tags Stats
// @Produce json
// @Success 200 {object} map[string]string
// @Router /quiz/stats [get]
func GetStats(w http.ResponseWriter, r *http.Request) {
	session, _ := SessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	// Delegate to the service layer
	statsMessage, err := services.GetStats(username)
	if err != nil {
		if err.Error() == "no stats available for user" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": statsMessage,
	})
}
