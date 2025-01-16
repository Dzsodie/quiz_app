package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Dzsodie/quiz_app/internal/services"
	"go.uber.org/zap"
)

type StatsHandler struct {
	StatsService services.IStatsService
	Logger       *zap.SugaredLogger
}

func NewStatsHandler(statsService services.IStatsService, logger *zap.Logger) *StatsHandler {
	return &StatsHandler{StatsService: statsService, Logger: logger.Sugar()}
}

// @Summary Get user statistics
// @Description Retrieve user performance statistics
// @Tags Stats
// @Produce json
// @Success 200 {object} map[string]string
// @Router /quiz/stats [get]
func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	session, _ := SessionStore.Get(r, "quiz-session")
	username, ok := session.Values["username"].(string)

	if !ok || username == "" {
		h.Logger.Warn("Failed to retrieve username from session")
		http.Error(w, `{"message":"Invalid session"}`, http.StatusUnauthorized)
		return
	}

	h.Logger.Info("Processing stats request", zap.String("username", username))

	statsMessage, err := h.StatsService.GetStats(username)
	if err != nil {
		if errors.Is(err, services.ErrNoStatsForUser) {
			h.Logger.Warn("No statistics for user", zap.String("username", username))
			http.Error(w, `{"message":"No stats available for user"}`, http.StatusBadRequest)
			return
		}
		h.Logger.Error("Failed to retrieve statistics", zap.String("username", username))
		http.Error(w, `{"message":"Internal server error"}`, http.StatusInternalServerError)
		return
	}
	h.Logger.Info("Stats retrieved successfully", zap.String("username", username))
	if err := json.NewEncoder(w).Encode(map[string]string{"message": statsMessage}); err != nil {
		h.Logger.Warn("Failed to encode score response: %v", err)
	}
}
