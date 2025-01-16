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
}

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
	logger := utils.GetLogger().Sugar()
	w.Header().Set("Content-Type", "application/json")
	allQuestions, err := h.QuizService.GetQuestions()
	if err != nil {
		logger.Error("Internal server error during get questions", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("Questions retrieved")
	if err := json.NewEncoder(w).Encode(allQuestions); err != nil {
		logger.Warn("Failed to encode quiz questions: %v", err)
	}

}

// @Summary Start the quiz
// @Description Initialize a new quiz session
// @Tags Quiz
// @Success 200 {object} map[string]string "message"
// @Router /quiz/start [post]
func (h *QuizHandler) StartQuiz(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger().Sugar()
	session, _ := SessionStore.Get(r, "quiz-session")
	username, ok := session.Values["username"].(string)
	if !ok || username == "" {
		logger.Warn("No username found in session")
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	if err := h.QuizService.StartQuiz(username); err != nil {
		logger.Error("Failed to start quiz", zap.String("username", username), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("Quiz started", zap.String("username", username))
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":        "quiz started",
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
	logger := utils.GetLogger().Sugar()
	session, _ := SessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	question, err := h.QuizService.GetNextQuestion(username)
	if err != nil {
		if err.Error() == "no more questions" {
			logger.Info("Quiz complete", zap.String("username", username))
			if err := json.NewEncoder(w).Encode(map[string]string{
				"status":           "quiz complete",
				"results_endpoint": "/quiz/results",
			}); err != nil {
				logger.Warn("Failed to encode response: %v", err)
			}

			return
		}
		logger.Error("Internal server error during get next question", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	logger.Info("Next question retrieved", zap.String("username", username))
	if err := json.NewEncoder(w).Encode(question); err != nil {
		logger.Warn("Failed to encode question: %v", err)
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
	logger := utils.GetLogger().Sugar()
	var payload models.AnswerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		logger.Warn("Invalid input for submit answer", zap.Error(err))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	session, _ := SessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	allQuestions, err := h.QuizService.GetQuestions()
	if err != nil {
		logger.Error("Internal server error during get questions", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := utils.ValidateAnswerPayload(payload.QuestionIndex, payload.Answer, allQuestions); err != nil {
		logger.Warn("Validation failed for answer payload", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	correct, err := h.QuizService.SubmitAnswer(username, payload.QuestionIndex, payload.Answer)
	if err != nil {
		if err.Error() == "invalid question index" {
			logger.Warn("Invalid question index", zap.Error(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		logger.Error("Internal server error during submit answer", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if correct {
		logger.Info("Correct answer submitted", zap.String("username", username), zap.Int("questionIndex", payload.QuestionIndex))
		if err := json.NewEncoder(w).Encode(map[string]string{"message": "Correct answer"}); err != nil {
			logger.Warn("Failed to encode correct answer response: %v", err)
		}

	} else {
		logger.Info("Wrong answer submitted", zap.String("username", username), zap.Int("questionIndex", payload.QuestionIndex))
		if err := json.NewEncoder(w).Encode(map[string]string{"message": "Wrong answer"}); err != nil {
			logger.Warn("Failed to encode wrong answer response: %v", err)
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
	logger := utils.GetLogger().Sugar()
	session, _ := SessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	score, err := h.QuizService.GetResults(username) // Delegate to the service
	if err != nil {
		if err.Error() == "quiz not started" {
			logger.Warn("Quiz not started", zap.String("username", username))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		logger.Error("Internal server error during get results", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	logger.Info("Quiz results retrieved", zap.String("username", username), zap.Int("score", score))
	if err := json.NewEncoder(w).Encode(map[string]int{"score": score}); err != nil {
		logger.Warn("Failed to encode score response: %v", err)
	}
}
