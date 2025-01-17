package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func StartQuizCLI(apiBaseURL string) {
	reader := bufio.NewReader(os.Stdin)
	client := &http.Client{}

	printWelcomeMessage()

	_, sessionToken := handleAuthentication(reader, client, apiBaseURL)
	startQuizLoop(reader, client, apiBaseURL, sessionToken)
}

func printWelcomeMessage() {
	fmt.Println("Welcome to the CLI Quiz App!")
	fmt.Println("This is a fun and engaging quiz game. You will be asked multiple-choice questions.")
	fmt.Println("Try your best to answer correctly and see how well you score!")
}

func handleAuthentication(reader *bufio.Reader, client *http.Client, apiBaseURL string) (string, string) {
	var username, sessionToken string

	for {
		fmt.Println("Do you want to (1) Register or (2) Login?")
		fmt.Print("Enter 1 or 2: ")
		choice := readInput(reader)

		if choice == "1" {
			registerUser(reader, client, apiBaseURL)
			continue
		} else if choice == "2" {
			username, sessionToken = loginUser(reader, client, apiBaseURL)
			break
		} else {
			fmt.Println("Invalid choice. Please enter 1 or 2.")
		}
	}
	return username, sessionToken
}

func registerUser(reader *bufio.Reader, client *http.Client, apiBaseURL string) {
	fmt.Println("Register a new account.")
	fmt.Print("Enter a username: ")
	username := readInput(reader)

	fmt.Print("Enter a password: ")
	password := readInput(reader)

	registerPayload := map[string]string{
		"username": username,
		"password": password,
	}
	registerBody, _ := json.Marshal(registerPayload)
	registerResp, err := client.Post(apiBaseURL+"/register", "application/json", bytes.NewBuffer(registerBody))
	if err != nil || registerResp.StatusCode != http.StatusCreated {
		fmt.Println("Registration failed. Try again.")
		return
	}
	fmt.Println("Registration successful! Please login to continue.")
}

func loginUser(reader *bufio.Reader, client *http.Client, apiBaseURL string) (string, string) {
	fmt.Print("Enter your username: ")
	username := readInput(reader)

	fmt.Print("Enter your password: ")
	password := readInput(reader)

	authPayload := map[string]string{
		"username": username,
		"password": password,
	}
	authBody, _ := json.Marshal(authPayload)
	authResp, err := client.Post(apiBaseURL+"/login", "application/json", bytes.NewBuffer(authBody))
	if err != nil || authResp.StatusCode != http.StatusOK {
		if authResp != nil && authResp.StatusCode == http.StatusUnauthorized {
			fmt.Println("Invalid username or password. Try again.")
		} else {
			fmt.Printf("Login failed: %v\n", err)
		}
		return "", ""
	}

	sessionToken := extractSessionToken(authResp)
	fmt.Println("Login successful!")
	return username, sessionToken
}

func extractSessionToken(resp *http.Response) string {
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "quiz-session" {
			return cookie.Value
		}
	}
	return ""
}

func startQuizLoop(reader *bufio.Reader, client *http.Client, apiBaseURL, sessionToken string) {
	for {
		startQuiz(client, apiBaseURL, sessionToken)
		if !playQuiz(reader, client, apiBaseURL, sessionToken) {
			break
		}
	}
}

func startQuiz(client *http.Client, apiBaseURL, sessionToken string) {
	req, _ := http.NewRequest(http.MethodPost, apiBaseURL+"/quiz/start", nil)
	req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
	startResp, err := client.Do(req)
	if err != nil || startResp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to start quiz: %v\n", err)
		return
	}
	fmt.Println("Quiz started! Answer the following questions:")
}

func playQuiz(reader *bufio.Reader, client *http.Client, apiBaseURL, sessionToken string) bool {
	for {
		question, finished := fetchNextQuestion(client, apiBaseURL, sessionToken)
		if finished {
			fetchResultsAndStats(client, apiBaseURL, sessionToken)
			return askPlayAgain(reader)
		}
		if question == nil {
			fmt.Println("An error occurred while fetching the next question.")
			return false
		}
		processQuestion(reader, client, apiBaseURL, sessionToken, question)
	}
}

func fetchNextQuestion(client *http.Client, apiBaseURL, sessionToken string) (*struct {
	QuestionID int      `json:"question_id"`
	Question   string   `json:"question"`
	Options    []string `json:"options"`
}, bool) {
	req, _ := http.NewRequest(http.MethodGet, apiBaseURL+"/quiz/next", nil)
	req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to fetch the next question. Please try again later.")
		return nil, false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var question struct {
			QuestionID int      `json:"question_id"`
			Question   string   `json:"question"`
			Options    []string `json:"options"`
		}
		body, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &question)
		return &question, false
	} else if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusGone {
		// Handle quiz completion
		return nil, true
	}

	fmt.Println("Unexpected response while fetching the next question.")
	return nil, false
}

func processQuestion(reader *bufio.Reader, client *http.Client, apiBaseURL, sessionToken string, question *struct {
	QuestionID int      `json:"question_id"`
	Question   string   `json:"question"`
	Options    []string `json:"options"`
}) {
	fmt.Printf("\nQuestion: %s\n", question.Question)
	for i, option := range question.Options {
		fmt.Printf("%d. %s\n", i+1, option)
	}

	answer := getAnswer(reader, len(question.Options))
	submitAnswer(client, apiBaseURL, sessionToken, question.QuestionID, answer)
}

func submitAnswer(client *http.Client, apiBaseURL, sessionToken string, questionID, answer int) {
	answerPayload := map[string]int{
		"QuestionIndex": questionID,
		"Answer":        answer,
	}
	answerBody, _ := json.Marshal(answerPayload)
	req, _ := http.NewRequest(http.MethodPost, apiBaseURL+"/quiz/submit", bytes.NewBuffer(answerBody))
	req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to submit answer: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result map[string]string
		body, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &result)
		fmt.Println(result["message"])
	} else {
		fmt.Println("Failed to process your answer.")
	}
}

func getAnswer(reader *bufio.Reader, optionsCount int) int {
	for {
		fmt.Print("Enter the number of your answer: ")
		var answer int
		_, err := fmt.Scanf("%d", &answer)
		if err != nil || answer < 1 || answer > optionsCount {
			fmt.Println("Invalid input. Please enter a valid option number.")
			if _, err := reader.ReadString('\n'); err != nil {
				fmt.Printf("Error reading input: %v\n", err)
			}
			continue
		}
		return answer - 1 // Convert to zero-based index
	}
}

func fetchResultsAndStats(client *http.Client, apiBaseURL, sessionToken string) {
	fetchResults(client, apiBaseURL, sessionToken)
	fetchStats(client, apiBaseURL, sessionToken)
}

func fetchResults(client *http.Client, apiBaseURL, sessionToken string) {
	req, _ := http.NewRequest(http.MethodGet, apiBaseURL+"/quiz/results", nil)
	req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to retrieve results: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var score struct {
		Score int `json:"score"`
	}
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &score)
	fmt.Printf("\nYour final score: %d\n", score.Score)
}

func fetchStats(client *http.Client, apiBaseURL, sessionToken string) {
	req, _ := http.NewRequest(http.MethodGet, apiBaseURL+"/quiz/stats", nil)
	req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to retrieve stats: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var stats map[string]string
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &stats)
	fmt.Printf("Your stats: %v\n", stats)
}

func askPlayAgain(reader *bufio.Reader) bool {
	fmt.Println("Do you want to play again? (yes/no)")
	response := readInput(reader)
	return strings.TrimSpace(strings.ToLower(response)) == "yes"
}

func readInput(reader *bufio.Reader) string {
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
