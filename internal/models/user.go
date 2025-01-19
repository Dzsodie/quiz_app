package models

type User struct {
	UserID     string  `json:"userID"`
	Username   string  `json:"username"`
	Password   string  `json:"password"`
	Progress   []int   `json:"progress"`
	Score      int     `json:"score"`
	QuizTaken  int     `json:"quizTaken"`
	Percentage float64 `json:"percentage"`
}
