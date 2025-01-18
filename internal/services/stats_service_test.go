package services

import (
	"testing"

	"github.com/Dzsodie/quiz_app/internal/database"
	"github.com/stretchr/testify/assert"
)

func TestStatsServiceGetStats(t *testing.T) {
	db := database.NewMemoryDB()
	db.AddUser(database.User{Username: "testuser1", Score: 10})
	db.AddUser(database.User{Username: "testuser2", Score: 20})
	db.AddUser(database.User{Username: "testuser3", Score: 15})

	service := NewStatsService(db)

	t.Run("Valid user stats retrieval", func(t *testing.T) {
		_, stats, err := service.GetStats("testuser1")

		assert.NoError(t, err, "expected no error for a valid user")
		assert.Contains(t, stats, "Your score is 10", "expected stats to include the correct score")
		assert.Contains(t, stats, "better than other users", "expected stats to include relative performance")
	})

	t.Run("Non-existent user", func(t *testing.T) {
		_, _, err := service.GetStats("nonexistent")

		assert.Error(t, err, "expected error for a non-existent user")
		assert.Equal(t, ErrNoStatsForUser, err, "unexpected error message for non-existent user")
	})
}
