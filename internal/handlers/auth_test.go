package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) GetUserID(username string) (string, error) {
	args := m.Called(username)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) RegisterUser(username, password string) error {
	args := m.Called(username, password)
	return args.Error(0)
}

func (m *MockAuthService) AuthenticateUser(username, password string) error {
	args := m.Called(username, password)
	return args.Error(0)
}

func TestRegisterUserHandler(t *testing.T) {
	mockService := new(MockAuthService)
	authHandler := NewAuthHandler(mockService)

	tests := []struct {
		name            string
		input           models.User
		mockReturnErr   error
		mockUserID      string
		mockUserIDError error
		expectedStatus  int
		expectedBody    string
	}{
		{
			name:            "Valid user registration",
			input:           models.User{Username: "testuser", Password: "Password123!"},
			mockReturnErr:   nil,
			mockUserID:      "12345",
			mockUserIDError: nil,
			expectedStatus:  http.StatusCreated,
			expectedBody:    `{"message":"User registered successfully","userID":"12345"}`,
		},
		{
			name:           "Invalid input - empty body",
			input:          models.User{},
			mockReturnErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message":"Invalid input"}`,
		},
		{
			name:           "User already exists",
			input:          models.User{Username: "testuser", Password: "Password123!"},
			mockReturnErr:  errors.New("user already exists"),
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"message":"User already exists"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input.Username != "" || tt.input.Password != "" {
				mockService.On("RegisterUser", tt.input.Username, tt.input.Password).Return(tt.mockReturnErr).Once()
			}
			if tt.mockReturnErr == nil && tt.mockUserID != "" {
				mockService.On("GetUserID", tt.input.Username).Return(tt.mockUserID, tt.mockUserIDError).Once()
			}

			body, _ := json.Marshal(tt.input)
			req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			authHandler.RegisterUser(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
