package services

import (
	"errors"
	"fmt"
	"sort"

	"github.com/Dzsodie/quiz_app/internal/database"
	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/Dzsodie/quiz_app/internal/utils"
	"go.uber.org/zap"
)

type StatsService struct {
	DB *database.MemoryDB
}

var (
	ErrNoStatsForUser = errors.New("no stats available for user")
)

// NewStatsService creates a new StatsService instance.
func NewStatsService(db *database.MemoryDB) *StatsService {
	return &StatsService{DB: db}
}

// GetStats calculates and returns a user's stats as a string.
func (s *StatsService) GetStats(username string) ([]models.User, string, error) {
	logger := utils.GetLogger().Sugar()
	logger.Info("Fetching stats for user", zap.String("username", username))

	user, err := s.DB.GetUser(username)
	if err != nil {
		logger.Warn("Stats not available for user", zap.String("username", username))
		return nil, "", ErrNoStatsForUser
	}

	allUsers := s.DB.GetAllUsers()
	logger.Debug("All users", zap.Any("users", allUsers))
	allScores := make([]int, len(allUsers))
	logger.Debug("All scores", zap.Any("scores", allScores))
	for i, user := range allUsers {
		allScores[i] = user.Score
	}
	sort.Ints(allScores)

	betterScores := 0
	for _, score := range allScores {
		if user.Score > score {
			betterScores++
		}
	}
	logger.Debug("Better scores", zap.Int("better_scores", betterScores))
	percentage := (float64(betterScores) / float64(len(allScores))) * 100
	logger.Debug("Percentage", zap.Float64("percentage", percentage))
	// Create the response message
	message := fmt.Sprintf(
		"Your score is %d and that is %.2f%% better than other users' scores.",
		user.Score, percentage,
	)

	logger.Info("Stats calculated successfully",
		zap.String("username", username),
		zap.Int("score", user.Score),
		zap.Float64("better_than_percentage", percentage),
	)

	// Convert allUsers to []models.User
	modelUsers := make([]models.User, len(allUsers))
	for i, user := range allUsers {
		modelUsers[i] = models.User{
			Username: user.Username,
			Score:    user.Score,
		}
	}

	return modelUsers, message, nil
}

var _ IStatsService = &StatsService{}
