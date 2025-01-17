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

	fmt.Println("Welcome to the CLI Quiz App!")
	fmt.Println("This is a fun and engaging quiz game. You will be asked multiple-choice questions.")
	fmt.Println("Try your best to answer correctly and see how well you score!")

	var username, sessionToken string

	for {
		fmt.Println("Do you want to (1) Register or (2) Login?")
		fmt.Print("Enter 1 or 2: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		if choice == "1" {
			// Registration
			fmt.Println("Register a new account.")
			fmt.Print("Enter a username: ")
			username, _ = reader.ReadString('\n')
			username = strings.TrimSpace(username)

			fmt.Print("Enter a password: ")
			password, _ := reader.ReadString('\n')
			password = strings.TrimSpace(password)

			registerPayload := map[string]string{
				"username": username,
				"password": password,
			}
			registerBody, _ := json.Marshal(registerPayload)
			registerResp, err := client.Post(apiBaseURL+"/register", "application/json", bytes.NewBuffer(registerBody))
			if err != nil || registerResp.StatusCode != http.StatusCreated {
				fmt.Println("Registration failed. Try again.")
				continue
			}
			fmt.Println("Registration successful! Please login to continue.")
			continue
		} else if choice == "2" {
			// Login
			fmt.Print("Enter your username: ")
			username, _ = reader.ReadString('\n')
			username = strings.TrimSpace(username)

			fmt.Print("Enter your password: ")
			password, _ := reader.ReadString('\n')
			password = strings.TrimSpace(password)

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
				continue
			}

			// Save session token (cookie)
			cookies := authResp.Cookies()
			for _, cookie := range cookies {
				if cookie.Name == "quiz-session" {
					sessionToken = cookie.Value
					break
				}
			}
			fmt.Println("Login successful!")
			break
		} else {
			fmt.Println("Invalid choice. Please enter 1 or 2.")
		}
	}

	for {
		// Start Quiz
		req, _ := http.NewRequest(http.MethodPost, apiBaseURL+"/quiz/start", nil)
		req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
		startResp, err := client.Do(req)
		if err != nil || startResp.StatusCode != http.StatusOK {
			fmt.Printf("Failed to start quiz: %v\n", err)
			return
		}
		fmt.Println("Quiz started! Answer the following questions:")

		// Quiz Loop
		for {
			req, _ := http.NewRequest(http.MethodGet, apiBaseURL+"/quiz/next", nil)
			req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
			questionResp, err := client.Do(req)
			if err != nil {
				fmt.Printf("Failed to retrieve question: %v\n", err)
				return
			}

			if questionResp.StatusCode == http.StatusOK {
				var question struct {
					QuestionID int      `json:"question_id"`
					Question   string   `json:"question"`
					Options    []string `json:"options"`
				}
				body, _ := io.ReadAll(questionResp.Body)
				_ = json.Unmarshal(body, &question)

				fmt.Printf("\nQuestion: %s\n", question.Question)
				for i, option := range question.Options {
					fmt.Printf("%d. %s\n", i+1, option)
				}

				var answer int
				for {
					fmt.Print("Enter the number of your answer: ")
					_, err := fmt.Scanf("%d", &answer)
					if err != nil || answer < 1 || answer > len(question.Options) {
						fmt.Println("Invalid input. Please enter a valid option number.")
						reader.ReadString('\n')
						continue
					}
					break
				}

				answerPayload := map[string]int{
					"QuestionIndex": question.QuestionID,
					"Answer":        answer - 1,
				}
				answerBody, _ := json.Marshal(answerPayload)

				req, _ := http.NewRequest(http.MethodPost, apiBaseURL+"/quiz/submit", bytes.NewBuffer(answerBody))
				req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
				submitResp, err := client.Do(req)
				if err != nil {
					fmt.Printf("Failed to submit answer: %v\n", err)
					return
				}

				if submitResp.StatusCode == http.StatusOK {
					var result map[string]string
					body, _ := io.ReadAll(submitResp.Body)
					_ = json.Unmarshal(body, &result)
					fmt.Println(result["message"])
				} else {
					fmt.Println("Failed to process your answer.")
				}
			} else if questionResp.StatusCode == http.StatusNoContent {
				fmt.Println("Quiz complete!")
				break
			} else {
				fmt.Println("An error occurred while fetching the next question.")
				break
			}
		}

		// Fetch Results
		req, _ = http.NewRequest(http.MethodGet, apiBaseURL+"/quiz/results", nil)
		req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
		resultsResp, err := client.Do(req)
		if err != nil || resultsResp.StatusCode != http.StatusOK {
			fmt.Printf("Failed to retrieve results: %v\n", err)
			return
		}

		var score struct {
			Score int `json:"score"`
		}
		body, _ := io.ReadAll(resultsResp.Body)
		_ = json.Unmarshal(body, &score)
		fmt.Printf("\nYour final score: %d\n", score.Score)

		// Fetch Stats
		req, _ = http.NewRequest(http.MethodGet, apiBaseURL+"/quiz/stats", nil)
		req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
		statsResp, err := client.Do(req)
		if err != nil || statsResp.StatusCode != http.StatusOK {
			fmt.Printf("Failed to retrieve stats: %v\n", err)
			return
		}

		var stats map[string]string
		body, _ = io.ReadAll(statsResp.Body)
		_ = json.Unmarshal(body, &stats)
		fmt.Printf("Your stats: %v\n", stats)

		// Play Again?
		fmt.Println("Do you want to play again? (yes/no)")
		response, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(response)) != "yes" {
			fmt.Println("Thank you for playing! Goodbye.")
			break
		}
	}
}
