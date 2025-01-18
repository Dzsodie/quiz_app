package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Dzsodie/quiz_app/internal/services"
	"github.com/Dzsodie/quiz_app/internal/utils"
	"go.uber.org/zap"
)

type StatsHandler struct {
	StatsService services.IStatsService
}

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
	logger := utils.GetLogger().Sugar()
	session, _ := SessionStore.Get(r, "quiz-session")
	username, ok := session.Values["username"].(string)

	if !ok || username == "" {
		logger.Warn("Failed to retrieve username from session")
		http.Error(w, `{"message":"Invalid session"}`, http.StatusUnauthorized)
		return
	}

	logger.Info("Processing stats request", zap.String("username", username))

	_, statsMessage, err := h.StatsService.GetStats(username)
	if err != nil {
		if errors.Is(err, services.ErrNoStatsForUser) {
			logger.Warn("No statistics for user", zap.String("username", username))
			http.Error(w, `{"message":"No stats available for user"}`, http.StatusBadRequest)
			return
		}
		logger.Error("Failed to retrieve statistics", zap.String("username", username))
		http.Error(w, `{"message":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	logger.Info("Stats retrieved successfully", zap.String("username", username))
	response := map[string]string{"message": statsMessage}
	statsJSON, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal stats message", zap.Error(err))
		http.Error(w, `{"message":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(statsJSON); err != nil {
		logger.Warn("Failed to write stats response", zap.Error(err))
	}
}
