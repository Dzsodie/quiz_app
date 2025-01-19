package handlers

import (
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
