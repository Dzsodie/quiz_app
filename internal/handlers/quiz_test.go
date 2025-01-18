package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockQuizService is a mock implementation of the IQuizService interface.
type MockQuizService struct {
	mock.Mock
}

func (m *MockQuizService) GetQuestions() ([]models.Question, error) {
	args := m.Called()
	return args.Get(0).([]models.Question), args.Error(1)
}

func (m *MockQuizService) LoadQuestions(questions []models.Question) {
	m.Called(questions)
}

func (m *MockQuizService) StartQuiz(username string) error {
	args := m.Called(username)
	return args.Error(0)
}

func (m *MockQuizService) GetNextQuestion(username string) (*models.Question, error) {
	args := m.Called(username)
	return args.Get(0).(*models.Question), args.Error(1)
}

func (m *MockQuizService) SubmitAnswer(username string, questionIndex int, answer int) (bool, error) {
	args := m.Called(username, questionIndex, answer)
	return args.Bool(0), args.Error(1)
}

func (m *MockQuizService) GetResults(username string) (int, error) {
	args := m.Called(username)
	return args.Int(0), args.Error(1)
}

func (m *MockQuizService) GetStats(username string) ([]models.User, string, error) {
	args := m.Called(username)
	return args.Get(0).([]models.User), args.String(1), args.Error(2)
}

func TestGetQuestions(t *testing.T) {
	mockService := new(MockQuizService)
	handler := NewQuizHandler(mockService)

	expectedQuestions := []models.Question{{QuestionID: 1, Question: "What is Go?"}}
	mockService.On("GetQuestions").Return(expectedQuestions, nil)

	req := httptest.NewRequest(http.MethodGet, "/questions", nil)
	rr := httptest.NewRecorder()

	handler.GetQuestions(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var actualQuestions []models.Question
	err := json.Unmarshal(rr.Body.Bytes(), &actualQuestions)
	assert.NoError(t, err)
	assert.Equal(t, expectedQuestions, actualQuestions)
	mockService.AssertExpectations(t)
}

func TestStartQuiz(t *testing.T) {
	mockService := new(MockQuizService)
	handler := NewQuizHandler(mockService)

	mockService.On("StartQuiz", "test_user").Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/quiz/start", nil)
	req.Header.Set("Content-Type", "application/json")

	// Mock session
	req = withMockSession(req, "test_user")

	rr := httptest.NewRecorder()
	handler.StartQuiz(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}

func TestNextQuestion(t *testing.T) {
	mockService := new(MockQuizService)
	handler := NewQuizHandler(mockService)

	expectedQuestion := &models.Question{QuestionID: 1, Question: "What is Go?"}
	mockService.On("GetNextQuestion", "test_user").Return(expectedQuestion, nil)

	req := httptest.NewRequest(http.MethodGet, "/quiz/next", nil)
	req = withMockSession(req, "test_user")

	rr := httptest.NewRecorder()
	handler.NextQuestion(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var actualQuestion models.Question
	err := json.Unmarshal(rr.Body.Bytes(), &actualQuestion)
	assert.NoError(t, err)
	assert.Equal(t, *expectedQuestion, actualQuestion)
	mockService.AssertExpectations(t)
}

func TestSubmitAnswer(t *testing.T) {
	mockService := new(MockQuizService)
	handler := NewQuizHandler(mockService)

	// Mock data for GetQuestions
	expectedQuestions := []models.Question{
		{
			QuestionID: 1,
			Question:   "What is Go?",
			Options:    []string{"A programming language", "A board game"},
			Answer:     0,
		},
	}
	mockService.On("GetQuestions").Return(expectedQuestions, nil).Once()
	mockService.On("SubmitAnswer", "test_user", 0, 0).Return(true, nil).Once()

	payload := models.AnswerPayload{
		QuestionIndex: 0,
		Answer:        0,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/quiz/submit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withMockSession(req, "test_user")

	rr := httptest.NewRecorder()
	handler.SubmitAnswer(rr, req)

	if rr.Code != http.StatusOK {
		t.Logf("Response Body: %s", rr.Body.String())
	}

	assert.Equal(t, http.StatusOK, rr.Code, "Expected HTTP 200 OK")

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	assert.Equal(t, "Correct answer", response["message"])

	mockService.AssertExpectations(t)
}

func TestGetResults(t *testing.T) {
	mockService := new(MockQuizService)
	handler := NewQuizHandler(mockService)

	mockService.On("GetResults", "test_user").Return(10, nil)

	req := httptest.NewRequest(http.MethodGet, "/quiz/results", nil)
	req = withMockSession(req, "test_user")

	rr := httptest.NewRecorder()
	handler.GetResults(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var result map[string]int
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, 10, result["score"])
	mockService.AssertExpectations(t)
}

func withMockSession(req *http.Request, username string) *http.Request {
	rr := httptest.NewRecorder()
	session, _ := SessionStore.Get(req, "quiz-session")
	session.Values["username"] = username
	_ = session.Save(req, rr)
	return req
}
