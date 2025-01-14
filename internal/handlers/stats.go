package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Dzsodie/quiz_app/internal/services"
)

// StatsHandler handles statistics-related HTTP requests.
type StatsHandler struct {
	StatsService services.IStatsService
}

// NewStatsHandler creates a new instance of StatsHandler with the provided IStatsService implementation.
func NewStatsHandler(statsService services.IStatsService) *StatsHandler {
	return &StatsHandler{StatsService: statsService}
}

// @Summary Get user statistics
// @Description Retrieve user performance statistics
// @Tags Stats
// @Produce json
// @Success 200 {object} map[string]string
// @Router /quiz/stats [get]
func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	session, _ := SessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	// Delegate to the service layer
	statsMessage, err := h.StatsService.GetStats(username)
	if err != nil {
		if err.Error() == "no stats available for user" {
			http.Error(w, `{"message":"No stats available for user"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"message":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": statsMessage,
	})
}
