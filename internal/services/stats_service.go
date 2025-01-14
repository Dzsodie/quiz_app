package services

import (
	"fmt"
	"sort"
	"sync"
)

var statsMu sync.Mutex

// GetStats calculates the user's performance relative to others.
func GetStats(username string) (string, error) {
	statsMu.Lock()
	defer statsMu.Unlock()

	userScore, exists := userScores[username]
	if !exists {
		return "", fmt.Errorf("no stats available for user: %s", username)
	}

	allScores := make([]int, 0, len(userScores))
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
	return fmt.Sprintf("Your score is %d and that is %.2f%% better than other users' scores.", userScore, percentage), nil
}
