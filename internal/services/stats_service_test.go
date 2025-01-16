package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatsServiceGetStats(t *testing.T) {
	s := &StatsService{}

	userScores["testuser1"] = 10
	userScores["testuser2"] = 20
	userScores["testuser3"] = 15

	stats, err := s.GetStats("testuser1")
	assert.NoError(t, err, "expected no error when fetching stats for an existing user")
	assert.Contains(t, stats, "Your score is 10", "expected stats to include user's score")
	assert.Contains(t, stats, "better than other users", "expected stats to include relative performance")

	_, err = s.GetStats("nonexistent")
	assert.Error(t, err, "expected error when fetching stats for a non-existent user")
	assert.Equal(t, ErrNoStatsForUser, err, "unexpected error message")
}
