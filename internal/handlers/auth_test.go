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

// MockAuthService is a mock implementation of the IAuthService interface.
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
	// Create the mock service
	mockService := new(MockAuthService)

	// Create the AuthHandler with the mock service
	authHandler := NewAuthHandler(mockService)

	// Define test cases for RegisterUser
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
			expectedBody:   "Invalid input\n",
		},
		{
			name:           "User already exists",
			input:          models.User{Username: "testuser", Password: "password"},
			mockReturnErr:  errors.New("user already exists"),
			expectedStatus: http.StatusConflict,
			expectedBody:   "user already exists\n",
		},
	}

	// Run each test case for RegisterUser
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock's expected behavior
			mockService.On("RegisterUser", tt.input.Username, tt.input.Password).Return(tt.mockReturnErr)

			// Create a new request
			body, _ := json.Marshal(tt.input)
			req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			assert.NoError(t, err)

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			authHandler.RegisterUser(rr, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.JSONEq(t, tt.expectedBody, rr.Body.String())

			// Assert that the mock expectations were met
			mockService.AssertExpectations(t)
		})
	}
}

func TestLoginUserHandler(t *testing.T) {
	// Create the mock service
	mockService := new(MockAuthService)

	// Create the AuthHandler with the mock service
	authHandler := NewAuthHandler(mockService)

	// Define test cases for LoginUser
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
			expectedBody:   "Invalid input\n",
		},
		{
			name:           "Invalid credentials",
			input:          models.User{Username: "testuser", Password: "wrongpassword"},
			mockReturnErr:  errors.New("invalid username or password"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid username or password\n",
		},
	}

	// Run each test case for LoginUser
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock's expected behavior
			mockService.On("AuthenticateUser", tt.input.Username, tt.input.Password).Return(tt.mockReturnErr)

			// Create a new request
			body, _ := json.Marshal(tt.input)
			req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			assert.NoError(t, err)

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			authHandler.LoginUser(rr, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.JSONEq(t, tt.expectedBody, rr.Body.String())

			// Assert that the mock expectations were met
			mockService.AssertExpectations(t)
		})
	}
}
