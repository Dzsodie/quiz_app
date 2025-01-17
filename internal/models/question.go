package models

type Question struct {
	QuestionID int      `json:"question_id"`
	Question   string   `json:"question"`
	Options    []string `json:"options"`
	Answer     int      `json:"answer"`
}
