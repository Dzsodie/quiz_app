package services

import "github.com/Dzsodie/quiz_app/internal/models"

type IStatsService interface {
	GetStats(username string) ([]models.User, string, error)
}
