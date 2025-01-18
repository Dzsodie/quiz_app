package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Dzsodie/quiz_app/internal/database"
)

type Question struct {
	QuestionID int      `json:"question_id"`
	Question   string   `json:"question"`
	Options    []string `json:"options"`
}

type StartQuizCLIService struct {
	ApiBaseURL string
	HttpClient *http.Client
	DB         *database.MemoryDB
}

func NewStartQuizCLIService(apiBaseURL string, client *http.Client, db *database.MemoryDB) *StartQuizCLIService {
	return &StartQuizCLIService{ApiBaseURL: apiBaseURL, HttpClient: client, DB: db}
}

func (s *StartQuizCLIService) RegisterUser(username, password string) (string, error) {
	payload := map[string]string{"username": username, "password": password}
	body, _ := json.Marshal(payload)
	resp, err := s.HttpClient.Post(s.ApiBaseURL+"/register", "application/json", bytes.NewBuffer(body))
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registration failed")
	}
	token, err := extractSessionToken(resp)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *StartQuizCLIService) LoginUser(username, password string) (string, error) {
	payload := map[string]string{"username": username, "password": password}
	body, _ := json.Marshal(payload)
	resp, err := s.HttpClient.Post(s.ApiBaseURL+"/login", "application/json", bytes.NewBuffer(body))
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed")
	}
	token, err := extractSessionToken(resp)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *StartQuizCLIService) StartQuiz(sessionToken string) error {
	req, _ := http.NewRequest(http.MethodPost, s.ApiBaseURL+"/quiz/start", nil)
	req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
	resp, err := s.HttpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to start quiz")
	}
	return nil
}

func (s *StartQuizCLIService) GetNextQuestion(sessionToken string) (*Question, bool, error) {
	req, _ := http.NewRequest(http.MethodGet, s.ApiBaseURL+"/quiz/next", nil)
	req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch next question")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var question Question
		body, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &question)
		return &question, false, nil
	} else if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusGone {
		return nil, true, nil
	}

	return nil, false, fmt.Errorf("unexpected response")
}

func (s *StartQuizCLIService) SubmitAnswer(sessionToken string, questionID, answer int) (string, error) {
	payload := map[string]int{"QuestionIndex": questionID, "Answer": answer}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, s.ApiBaseURL+"/quiz/submit", bytes.NewBuffer(body))
	req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to submit answer")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result map[string]string
		body, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &result)
		return result["message"], nil
	}

	return "", fmt.Errorf("unexpected response")
}

func (s *StartQuizCLIService) FetchResults(sessionToken string) (int, error) {
	req, _ := http.NewRequest(http.MethodGet, s.ApiBaseURL+"/quiz/results", nil)
	req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch results")
	}
	defer resp.Body.Close()

	var score struct {
		Score int `json:"score"`
	}
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &score)
	return score.Score, nil
}

func (s *StartQuizCLIService) FetchStats(sessionToken string) (map[string]string, error) {
	req, _ := http.NewRequest(http.MethodGet, s.ApiBaseURL+"/quiz/stats", nil)
	req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stats")
	}
	defer resp.Body.Close()

	var stats map[string]string
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &stats)
	return stats, nil
}

func extractSessionToken(resp *http.Response) (string, error) {
	var result map[string]string
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}
	token, ok := result["session_token"]
	if !ok {
		return "", fmt.Errorf("session token not found")
	}
	return token, nil
}
