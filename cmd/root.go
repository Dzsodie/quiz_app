/*
Copyright © 2025 Zsuzsa Makara <dzsodie@gmail.com>
*/
package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var cliMode bool

var rootCmd = &cobra.Command{
	Use:   "quiz_app",
	Short: "A CLI app for quiz questions and answers",
	Long: `quiz_app is a small CLI application built in Go that
contains quiz questions and answers, it lets the user choose
one answer and then evaluates the answer while sharing stats.`,
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
	resp, err := http.Get("http://localhost:8080/quiz/stats")
	if err != nil {
		fmt.Printf("Error fetching stats: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed to fetch stats. Ensure you are logged in.")
		return
	}

	fmt.Println("Your current stats:")
	fmt.Println(resp.Body)
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

	resp, err := http.Post("http://localhost:8080/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Println("Login failed.")
		return false
	}
	fmt.Println("Login successful.")
	return true
}

func quizLoop() {
	for {
		resp, err := http.Get("http://localhost:8080/quiz/next")
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
		json.NewDecoder(resp.Body).Decode(&question)

		fmt.Printf("\nQuestion: %s\n", question["question"])
		options := question["options"].([]interface{})
		for i, option := range options {
			fmt.Printf("%d. %s\n", i+1, option)
		}

		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Enter your answer: ")
		scanner.Scan()
		answer := scanner.Text()

		answerData := map[string]interface{}{
			"question_index": question["question_id"],
			"answer":         answer,
		}
		jsonData, _ := json.Marshal(answerData)

		resp, err = http.Post("http://localhost:8080/quiz/submit", "application/json", bytes.NewBuffer(jsonData))
		if err != nil || resp.StatusCode != http.StatusOK {
			fmt.Println("Error submitting answer.")
			return
		}

		var response map[string]string
		json.NewDecoder(resp.Body).Decode(&response)
		fmt.Println(response["message"])
	}
}
