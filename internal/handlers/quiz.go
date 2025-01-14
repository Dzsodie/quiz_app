package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/Dzsodie/quiz_app/internal/services"
)

// QuizHandler handles quiz-related HTTP requests.
type QuizHandler struct {
	QuizService services.IQuizService
}

// NewQuizHandler creates a new instance of QuizHandler with the provided IQuizService implementation.
func NewQuizHandler(quizService services.IQuizService) *QuizHandler {
	return &QuizHandler{QuizService: quizService}
}

// @Summary Get all questions
// @Description Retrieve the list of questions
// @Tags Quiz
// @Produce json
// @Success 200 {array} models.Question
// @Router /questions [get]
func (h *QuizHandler) GetQuestions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	allQuestions := h.QuizService.GetQuestions()
	json.NewEncoder(w).Encode(allQuestions)
}

// @Summary Start the quiz
// @Description Initialize a new quiz session
// @Tags Quiz
// @Success 200 {object} map[string]string "message"
// @Router /quiz/start [post]
func (h *QuizHandler) StartQuiz(w http.ResponseWriter, r *http.Request) {
	session, _ := SessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	h.QuizService.StartQuiz(username) // Delegate to the service

	json.NewEncoder(w).Encode(map[string]string{
		"message":       "Quiz started",
		"next_endpoint": "/quiz/next",
	})
}

// @Summary Get the next question
// @Description Retrieve the next question for the quiz
// @Tags Quiz
// @Produce json
// @Success 200 {object} models.Question
// @Router /quiz/next [get]
func (h *QuizHandler) NextQuestion(w http.ResponseWriter, r *http.Request) {
	session, _ := SessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	question, err := h.QuizService.GetNextQuestion(username) // Delegate to the service
	if err != nil {
		if err.Error() == "no more questions" {
			json.NewEncoder(w).Encode(map[string]string{
				"message":          "Quiz complete",
				"results_endpoint": "/quiz/results",
			})
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(question)
}

// @Summary Submit an answer
// @Description Submit an answer to a question
// @Tags Quiz
// @Accept json
// @Produce json
// @Param payload body models.AnswerPayload true "Answer payload"
// @Success 200 {object} map[string]string "message"
// @Failure 400 {object} map[string]string "Invalid input"
// @Router /quiz/submit [post]
func (h *QuizHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	var payload models.AnswerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	session, _ := SessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	correct, err := h.QuizService.SubmitAnswer(username, payload.QuestionIndex, payload.Answer) // Delegate to the service
	if err != nil {
		if err.Error() == "invalid question index" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if correct {
		json.NewEncoder(w).Encode(map[string]string{"message": "Correct answer"})
	} else {
		json.NewEncoder(w).Encode(map[string]string{"message": "Wrong answer"})
	}
}

// @Summary Get quiz results
// @Description Retrieve the results of the quiz
// @Tags Quiz
// @Produce json
// @Success 200 {object} map[string]int "score"
// @Router /quiz/results [get]
func (h *QuizHandler) GetResults(w http.ResponseWriter, r *http.Request) {
	session, _ := SessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	score, err := h.QuizService.GetResults(username) // Delegate to the service
	if err != nil {
		if err.Error() == "quiz not started" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"score": score})
}
