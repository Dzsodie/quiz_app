package services

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

type StatsService struct{}

var (
	statsMu           sync.Mutex
	ErrNoStatsForUser = errors.New("no stats available for user")
)

// GetStats calculates the user's performance relative to others.
func (s *StatsService) GetStats(username string) (string, error) {
	statsMu.Lock()
	defer statsMu.Unlock()

	userScore, exists := userScores[username]
	if !exists {
		return "", ErrNoStatsForUser
	}

	// Collect all scores for ranking
	allScores := []int{}
	for _, score := range userScores {
		allScores = append(allScores, score)
	}
	sort.Ints(allScores)

	// Calculate relative performance
	betterScores := 0
	for _, score := range allScores {
		if userScore > score {
			betterScores++
		}
	}

	totalUsers := len(allScores)
	percentage := (float64(betterScores) / float64(totalUsers)) * 100
	return fmt.Sprintf("Your score is %d and that is %.2f%% better than other users' scores.", userScore, percentage), nil
}

var _ IStatsService = &StatsService{}
