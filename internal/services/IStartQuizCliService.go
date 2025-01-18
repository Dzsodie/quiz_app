package services

type IStartQuizCLIService interface {
	RegisterUser(username, password string) (string, error)
	LoginUser(username, password string) (string, error)
	StartQuiz(sessionToken string) error
	GetNextQuestion(sessionToken string) (*Question, bool, error)
	SubmitAnswer(sessionToken string, questionID, answer int) (string, error)
	FetchResults(sessionToken string) (int, error)
	FetchStats(sessionToken string) (map[string]string, error)
}
