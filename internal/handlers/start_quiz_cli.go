package handlers

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Dzsodie/quiz_app/internal/database"
	"github.com/Dzsodie/quiz_app/internal/services"
	"github.com/Dzsodie/quiz_app/internal/utils"
)

type StartQuizCLIHandler struct {
	Service services.IStartQuizCLIService
}

func NewStartQuizCliHandler(startQuizCliService services.IStartQuizCLIService) *StartQuizCLIHandler {
	return &StartQuizCLIHandler{Service: startQuizCliService}
}

func (h *StartQuizCLIHandler) StartQuizCLI(APIBaseURL string, db *database.MemoryDB) {
	logger := utils.GetLogger().Sugar()
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to the CLI Quiz App!")

	for {
		fmt.Println("Do you want to (1) Register or (2) Login?")
		fmt.Print("Enter 1 or 2: ")
		choice := readInput(reader)

		if choice == "1" {
			h.handleRegister(reader)
		} else if choice == "2" {
			sessionToken := h.handleLogin(reader)
			logger.Debug("Session token_start", "session_token: ", sessionToken)
			if sessionToken != "" {
				h.startQuizLoop(reader, sessionToken)
				logger.Debug("Session token>>", "session_token: ", sessionToken)
				break
			}
		} else {
			logger.Warn("Invalid choice", "choice", choice)
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}

func (h *StartQuizCLIHandler) handleRegister(reader *bufio.Reader) {
	logger := utils.GetLogger().Sugar()
	fmt.Println("Register a new account.")
	fmt.Print("Enter a username: ")
	username := readInput(reader)

	fmt.Print("Enter a password: ")
	password := readInput(reader)

	userID, err := h.Service.RegisterUser(username, password)
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
		logger.Error("User registration failed. ", "error: ", err, "username: ", username, "password: ", password)
		return
	}

	fmt.Printf("Registration successful! Your user ID is: %s\n", userID)
	logger.Info("User registered successfully! ", "username: ", username, "userID: ", userID)
}

func (h *StartQuizCLIHandler) handleLogin(reader *bufio.Reader) string {
	logger := utils.GetLogger().Sugar()
	fmt.Print("Enter your username: ")
	username := readInput(reader)

	fmt.Print("Enter your password: ")
	password := readInput(reader)

	sessionToken, err := h.Service.LoginUser(username, password)
	logger.Debug("Session token_login", "session_token: ", sessionToken)
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		logger.Error("Login failed", "error", err, "username", username, "password", password)
		return ""
	}

	fmt.Println("Login successful!")
	logger.Info("User logged in successfully! ", "username: ", username, "session_token: ", sessionToken)
	return sessionToken
}

func (h *StartQuizCLIHandler) startQuizLoop(reader *bufio.Reader, sessionToken string) {
	logger := utils.GetLogger().Sugar()
	for {
		err := h.Service.StartQuiz(sessionToken)
		logger.Debug("Session token_startloop_h", "session_token: ", sessionToken)
		if err != nil {
			fmt.Println("Failed to start quiz. Try again.")
			return
		}

		finished := false
		for !finished {
			question, isFinished, err := h.Service.GetNextQuestion(sessionToken)
			if err != nil {
				fmt.Println("Error fetching question. Try again.")
				return
			}

			if isFinished {
				h.showResultsAndStats(sessionToken)
				finished = true
				break
			}

			fmt.Printf("\nQuestion: %s\n", question.Question)
			logger.Debug("Session token_getnextquestion_h", "session_token: ", sessionToken)

			for i, option := range question.Options {
				fmt.Printf("%d. %s\n", i+1, option)
			}

			answer := h.getAnswer(len(question.Options))
			message, err := h.Service.SubmitAnswer(sessionToken, question.QuestionID, answer)
			if err != nil {
				fmt.Println("Failed to submit answer. Try again.")
				return
			}

			fmt.Println(message)
		}

		if !h.askPlayAgain(reader) {
			break
		}
	}
}

func (h *StartQuizCLIHandler) showResultsAndStats(sessionToken string) {
	score, err := h.Service.FetchResults(sessionToken)
	if err == nil {
		fmt.Printf("Your final score: %d\n", score)
	} else {
		fmt.Println("Failed to fetch results.")
	}

	stats, err := h.Service.FetchStats(sessionToken)
	if err == nil {
		fmt.Printf("Your stats: %v\n", stats)
	} else {
		fmt.Println("Failed to fetch stats.")
	}
}

func (h *StartQuizCLIHandler) getAnswer(optionsCount int) int {
	for {
		fmt.Print("Enter the number of your answer: ")
		var answer int
		_, err := fmt.Scanf("%d", &answer)
		if err != nil || answer < 1 || answer > optionsCount {
			fmt.Println("Invalid input. Please try again.")
			continue
		}
		return answer - 1
	}
}

func (h *StartQuizCLIHandler) askPlayAgain(reader *bufio.Reader) bool {
	fmt.Println("Do you want to play again? (yes/y/no/n)")
	response := strings.TrimSpace(strings.ToLower(readInput(reader)))
	return response == "yes" || response == "y"
}

func readInput(reader *bufio.Reader) string {
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
