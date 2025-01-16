package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/Dzsodie/quiz_app/internal/services"
	"github.com/Dzsodie/quiz_app/internal/utils"
	"go.uber.org/zap"
)

type QuizHandler struct {
	QuizService services.IQuizService
	Logger      *zap.SugaredLogger
}

func NewQuizHandler(quizService services.IQuizService, logger *zap.SugaredLogger) *QuizHandler {
	return &QuizHandler{QuizService: quizService, Logger: logger}
}

// @Summary Get all questions
// @Description Retrieve the list of questions
// @Tags Quiz
// @Produce json
// @Success 200 {array} models.Question
// @Router /questions [get]
func (h *QuizHandler) GetQuestions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	allQuestions, err := h.QuizService.GetQuestions()
	if err != nil {
		h.Logger.Error("Internal server error during get questions", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.Logger.Info("Questions retrieved")
	if err := json.NewEncoder(w).Encode(allQuestions); err != nil {
		h.Logger.Warn("Failed to encode quiz questions: %v", err)
	}

}

// @Summary Start the quiz
// @Description Initialize a new quiz session
// @Tags Quiz
// @Success 200 {object} map[string]string "message"
// @Router /quiz/start [post]
func (h *QuizHandler) StartQuiz(w http.ResponseWriter, r *http.Request) {
	session, _ := SessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	if err := h.QuizService.StartQuiz(username); err != nil {
		h.Logger.Error("Internal server error during start quiz", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.Logger.Info("Quiz started", zap.String("username", username))
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":        "quiz started",
		"next_endpoint": "/quiz/next",
	}); err != nil {
		h.Logger.Warn("Failed to encode response: %v", err)
	}
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

	question, err := h.QuizService.GetNextQuestion(username)
	if err != nil {
		if err.Error() == "no more questions" {
			h.Logger.Info("Quiz complete", zap.String("username", username))
			if err := json.NewEncoder(w).Encode(map[string]string{
				"status":           "quiz complete",
				"results_endpoint": "/quiz/results",
			}); err != nil {
				h.Logger.Warn("Failed to encode response: %v", err)
			}

			return
		}
		h.Logger.Error("Internal server error during get next question", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.Logger.Info("Next question retrieved", zap.String("username", username))
	if err := json.NewEncoder(w).Encode(question); err != nil {
		h.Logger.Warn("Failed to encode question: %v", err)
	}

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
		h.Logger.Warn("Invalid input for submit answer", zap.Error(err))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	session, _ := SessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	allQuestions, err := h.QuizService.GetQuestions()
	if err != nil {
		h.Logger.Error("Internal server error during get questions", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := utils.ValidateAnswerPayload(payload.QuestionIndex, payload.Answer, allQuestions); err != nil {
		h.Logger.Warn("Validation failed for answer payload", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	correct, err := h.QuizService.SubmitAnswer(username, payload.QuestionIndex, payload.Answer)
	if err != nil {
		if err.Error() == "invalid question index" {
			h.Logger.Warn("Invalid question index", zap.Error(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.Logger.Error("Internal server error during submit answer", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if correct {
		h.Logger.Info("Correct answer submitted", zap.String("username", username), zap.Int("questionIndex", payload.QuestionIndex))
		if err := json.NewEncoder(w).Encode(map[string]string{"message": "Correct answer"}); err != nil {
			h.Logger.Warn("Failed to encode correct answer response: %v", err)
		}

	} else {
		h.Logger.Info("Wrong answer submitted", zap.String("username", username), zap.Int("questionIndex", payload.QuestionIndex))
		if err := json.NewEncoder(w).Encode(map[string]string{"message": "Wrong answer"}); err != nil {
			h.Logger.Warn("Failed to encode wrong answer response: %v", err)
		}

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
			h.Logger.Warn("Quiz not started", zap.String("username", username))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.Logger.Error("Internal server error during get results", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.Logger.Info("Quiz results retrieved", zap.String("username", username), zap.Int("score", score))
	if err := json.NewEncoder(w).Encode(map[string]int{"score": score}); err != nil {
		h.Logger.Warn("Failed to encode score response: %v", err)
	}
}
