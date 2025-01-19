package services

import "github.com/Dzsodie/quiz_app/internal/models"

type IStartQuizCLIService interface {
	RegisterUser(username, password string) (string, error)
	LoginUser(username, password string) (string, error)
	StartQuiz(sessionToken string) error
	GetNextQuestion(sessionToken string) (*models.Question, bool, error)
	SubmitAnswer(sessionToken string, questionID, answer int) (string, error)
	FetchResults(sessionToken string) (int, error)
	FetchStats(sessionToken string) (map[string]string, error)
}
