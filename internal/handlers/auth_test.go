package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
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
		name           string
		input          models.User
		mockReturnErr  error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid user registration",
			input:          models.User{Username: "testuser", Password: "password"},
			mockReturnErr:  nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"message":"User registered successfully"}`,
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
			input:          models.User{Username: "testuser", Password: "password"},
			mockReturnErr:  errors.New("user already exists"),
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"message":"user already exists"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input.Username != "" || tt.input.Password != "" {
				mockService.On("RegisterUser", tt.input.Username, tt.input.Password).Return(tt.mockReturnErr).Once()
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

func TestLoginUserHandler(t *testing.T) {
	mockService := new(MockAuthService)
	authHandler := NewAuthHandler(mockService)

	SessionStore = sessions.NewCookieStore([]byte("test-secret"))

	tests := []struct {
		name           string
		input          models.User
		mockReturnErr  error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid login",
			input:          models.User{Username: "testuser", Password: "password"},
			mockReturnErr:  nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Login successful"}`,
		},
		{
			name:           "Invalid input - empty body",
			input:          models.User{},
			mockReturnErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message":"Invalid input"}`,
		},
		{
			name:           "Invalid credentials",
			input:          models.User{Username: "testuser", Password: "wrongpassword"},
			mockReturnErr:  errors.New("invalid username or password"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"message":"invalid username or password"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input.Username != "" || tt.input.Password != "" {
				mockService.On("AuthenticateUser", tt.input.Username, tt.input.Password).Return(tt.mockReturnErr).Once()
			}

			body, _ := json.Marshal(tt.input)
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			authHandler.LoginUser(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
