/*
Copyright © 2025 Zsuzsa Makara <dzsodie@gmail.com>
*/
package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Dzsodie/quiz_app/internal/database"
	"github.com/Dzsodie/quiz_app/internal/handlers"
	"github.com/Dzsodie/quiz_app/internal/services"
	"github.com/Dzsodie/quiz_app/internal/utils"
	"github.com/spf13/cobra"
)

var quizService *services.QuizService
var authService *services.AuthService

var rootCmd = &cobra.Command{
	Use:   "quiz_app",
	Short: "A CLI app for quiz questions and answers",
	Long: `quiz_app is a small CLI application built in Go that
	contains quiz questions and answers, it lets the user choose
	one answer and then evaluates the answer while sharing stats.`,
}

var quizCmd = &cobra.Command{
	Use:   "quiz",
	Short: "Initiate the quiz application",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to the Quiz App!")
		fmt.Println("The quiz contains 10 questions with multiple choices.")
		fmt.Println("Commands available:")
		fmt.Println("1. start - Start the quiz")
		fmt.Println("2. score - View your score and stats")
		fmt.Println("3. exit - Quit the quiz app")
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the quiz",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting the quiz...")

		questions, err := utils.ReadCSV("questions.csv")
		if err != nil {
			fmt.Printf("Error loading questions: %v\n", err)
			return
		}
		quizService.LoadQuestions(questions)

		DB := database.NewMemoryDB()
		quizHandler := handlers.NewQuizHandler(services.NewQuizService(DB))

		http.HandleFunc("/quiz/start", quizHandler.StartQuiz)
		fmt.Println("Quiz started and available at: http://localhost:8080/quiz/start")

		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Printf("Error starting HTTP server: %v\n", err)
		}
	},
}

var scoreCmd = &cobra.Command{
	Use:   "score",
	Short: "View your score and stats",
	Run: func(cmd *cobra.Command, args []string) {
		if authService == nil {
			fmt.Println("Authentication service not initialized.")
			return
		}
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}
		session, err := authService.GetSession(req)
		if err != nil {
			fmt.Printf("Error retrieving session: %v\n", err)
			return
		}
		username, ok := session.Values["username"].(string)
		if !ok {
			fmt.Println("Invalid session data.")
			return
		}
		stats, _, err := quizService.GetStats(username)
		if err != nil {
			fmt.Printf("Error retrieving stats: %v\n", err)
			return
		}
		fmt.Printf("Your current stats: %v\n", stats)
	},
}

var exitCmd = &cobra.Command{
	Use:   "exit",
	Short: "Quit the quiz app",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Exiting the Quiz App. Goodbye!")
		os.Exit(0)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	DB := database.NewMemoryDB()
	quizService = services.NewQuizService(DB)
	authService = services.NewAuthService(DB)

	rootCmd.AddCommand(quizCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(scoreCmd)
	rootCmd.AddCommand(exitCmd)
}
