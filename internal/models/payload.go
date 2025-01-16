package models

type AnswerPayload struct {
	QuestionIndex int `json:"question_index"`
	Answer        int `json:"answer"`
}
