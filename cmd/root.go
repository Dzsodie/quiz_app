package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var cliMode bool
var sessionCookie string

var rootCmd = &cobra.Command{
	Use:   "quiz_app",
	Short: "A CLI app for quiz questions and answers",
	Long: `quiz_app is a small CLI application built in Go that
contains quiz questions and answers. It lets the user choose
one answer and evaluates the answer while sharing stats.`,
	Run: func(cmd *cobra.Command, args []string) {
		if cliMode {
			runCLI()
		} else {
			fmt.Println("Run the application using --cli to start in CLI mode.")
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&cliMode, "cli", false, "Run the application in CLI mode")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCLI() {
	fmt.Println("quiz_app")
	fmt.Println("A CLI app for quiz questions and answers")
	fmt.Println(`quiz_app is a small CLI application built in Go that contains quiz questions 
and answers. It lets the user choose one answer and evaluates the answer while sharing stats.`)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("\nAvailable commands:")
		fmt.Println("1. start - Start the quiz")
		fmt.Println("2. score - View your score and stats")
		fmt.Println("3. exit - Quit the quiz app")
		fmt.Print("\nEnter your command: ")
		scanner.Scan()
		input := scanner.Text()

		switch input {
		case "start":
			startQuizCLI()
		case "score":
			viewStatsCLI()
		case "exit":
			fmt.Println("Exiting the Quiz App. Goodbye!")
			os.Exit(0)
		default:
			fmt.Println("Invalid command. Please try again.")
		}
	}
}

func startQuizCLI() {
	fmt.Println("Registering a new user...")
	username, password := getUserCredentials()
	if !registerUser(username, password) {
		fmt.Println("Registration failed. Please try again.")
		return
	}

	fmt.Println("Logging in...")
	if !loginUser(username, password) {
		fmt.Println("Login failed. Please try again.")
		return
	}

	fmt.Println("Quiz started! Answer the questions as they appear.")
	quizLoop()
}

func viewStatsCLI() {
	if sessionCookie == "" {
		fmt.Println("You must log in before viewing stats.")
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8080/quiz/stats", nil)
	if err != nil {
		fmt.Printf("Error creating stats request: %v\n", err)
		return
	}
	req.Header.Set("Cookie", sessionCookie)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching stats: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed to fetch stats. Ensure you are logged in.")
		return
	}

	var statsResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&statsResponse); err != nil {
		fmt.Printf("Error decoding stats response: %v\n", err)
		return
	}

	if message, ok := statsResponse["message"].(string); ok {
		fmt.Println(message)
	} else {
		fmt.Println("Failed to retrieve stats message.")
	}
}

func getUserCredentials() (string, string) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter username: ")
	scanner.Scan()
	username := scanner.Text()

	fmt.Print("Enter password: ")
	scanner.Scan()
	password := scanner.Text()

	return username, password
}

func registerUser(username, password string) bool {
	data := map[string]string{
		"username": username,
		"password": password,
	}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post("http://localhost:8080/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil || resp.StatusCode != http.StatusCreated {
		fmt.Println("Registration failed.")
		return false
	}
	fmt.Println("Registration successful.")
	return true
}

func loginUser(username, password string) bool {
	data := map[string]string{
		"username": username,
		"password": password,
	}
	jsonData, _ := json.Marshal(data)

	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://localhost:8080/login", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating login request: %v\n", err)
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Println("Login failed.")
		return false
	}
	defer resp.Body.Close()

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "quiz-session" {
			sessionCookie = fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
			break
		}
	}

	if sessionCookie == "" {
		fmt.Println("Failed to retrieve session cookie.")
		return false
	}

	fmt.Println("Login successful.")
	return true
}

func quizLoop() {
	if sessionCookie == "" {
		fmt.Println("You must log in before starting the quiz.")
		return
	}

	client := &http.Client{}
	for {
		req, err := http.NewRequest("GET", "http://localhost:8080/quiz/next", nil)
		if err != nil {
			fmt.Printf("Error creating next question request: %v\n", err)
			return
		}
		req.Header.Set("Cookie", sessionCookie)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Error retrieving next question: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusGone {
			fmt.Println("Quiz complete! View your results by using the score command.")
			return
		} else if resp.StatusCode != http.StatusOK {
			fmt.Println("Error fetching question.")
			return
		}

		var question map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&question); err != nil {
			fmt.Printf("Error decoding question: %v\n", err)
			return
		}

		fmt.Printf("\nQuestion: %s\n", question["question"])
		options := question["options"].([]interface{})
		for i, option := range options {
			fmt.Printf("%d. %s\n", i+1, option)
		}

		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Enter your answer: ")
		scanner.Scan()
		answer := scanner.Text()

		answerInt, err := strconv.Atoi(answer)
		if err != nil || answerInt < 0 || answerInt > len(options)+1 {
			fmt.Println("Invalid answer. Please enter a valid option number.")
			continue
		}

		questionID := int(question["question_id"].(float64))
		answerData := map[string]interface{}{
			"question_index": questionID - 1,
			"answer":         answerInt,
		}
		jsonData, _ := json.Marshal(answerData)

		req, err = http.NewRequest("POST", "http://localhost:8080/quiz/submit", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error creating submit answer request: %v\n", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", sessionCookie)

		resp, err = client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			fmt.Println("Error submitting answer.")
			return
		}

		var response map[string]string
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			fmt.Printf("Error decoding answer response: %v\n", err)
			return
		}
		fmt.Println(response["message"])
	}
}
