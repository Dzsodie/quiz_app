package models

// AnswerPayload represents the data sent when submitting an answer.
type AnswerPayload struct {
	QuestionIndex int `json:"question_index"`
	Answer        int `json:"answer"`
}
