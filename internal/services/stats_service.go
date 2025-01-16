package services

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/Dzsodie/quiz_app/internal/utils"
	"go.uber.org/zap"
)

type StatsService struct{}

var (
	statsMu           sync.Mutex
	ErrNoStatsForUser = errors.New("no stats available for user")
)

func NewStatsService() *StatsService {
	return &StatsService{}
}

func (s *StatsService) GetStats(username string) (string, error) {
	logger := utils.GetLogger().Sugar()
	statsMu.Lock()
	defer statsMu.Unlock()

	logger.Info("Fetching stats for user", zap.String("username", username))

	userScore, exists := userScores[username]
	if !exists {
		logger.Warn("Stats not available for user", zap.String("username", username))
		return "", ErrNoStatsForUser
	}

	allScores := []int{}
	for _, score := range userScores {
		allScores = append(allScores, score)
	}
	sort.Ints(allScores)

	betterScores := 0
	for _, score := range allScores {
		if userScore > score {
			betterScores++
		}
	}

	totalUsers := len(allScores)
	percentage := (float64(betterScores) / float64(totalUsers)) * 100
	message := fmt.Sprintf("Your score is %d and that is %.2f%% better than other users' scores.", userScore, percentage)

	logger.Info("Stats calculated successfully",
		zap.String("username", username),
		zap.Int("score", userScore),
		zap.Float64("better_than_percentage", percentage),
	)

	return message, nil
}

var _ IStatsService = &StatsService{}
