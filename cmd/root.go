/*
Copyright © 2025 Zsuzsa Makara <dzsodie@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/Dzsodie/quiz_app/internal/handlers"
	"github.com/Dzsodie/quiz_app/internal/services"
	"github.com/Dzsodie/quiz_app/internal/utils"
	"github.com/spf13/cobra"
)

var quizService = &services.QuizService{}
var authService = &services.AuthService{}
var statsService = &services.StatsService{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "quiz_app",
	Short: "A CLI app for quiz questions and answers",
	Long: `quiz_app is a small CLI application built in Go that
	contains quiz questions and answers, it lets the user choose
	one answer and then evaluates the answer while sharing stats.`,
}

// quizCmd displays how-to information for the quiz app
var quizCmd = &cobra.Command{
	Use:   "quiz",
	Short: "Initiate the quiz application",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to the Quiz App!")
		fmt.Println("The quiz contains of 10 questions and 3-3 answers for each, make your choice of the answer by typing the number.")
		fmt.Println("There is a timer, try to answer all 10 questions within the given timeframe.")
		fmt.Println("The application is controllable with command line commands.")
		fmt.Println("Available commands:")
		fmt.Println("1. start - Start the quiz")
		fmt.Println("2. score - View your score and stats")
		fmt.Println("3. exit - Quit the quiz app")
	},
}

// startCmd starts the quiz
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
		apiBaseURL := "http://localhost:8080/api"
		handlers.StartQuizCLI(apiBaseURL)
	},
}

// scoreCmd shows the score and stats
var scoreCmd = &cobra.Command{
	Use:   "score",
	Short: "View your score and stats",
	Run: func(cmd *cobra.Command, args []string) {
		session, err := authService.GetSession()
		if err != nil {
			fmt.Printf("Error retrieving session: %v\n", err)
			return
		}
		username := session.Values["username"].(string)
		stats, err := statsService.GetStats(username)
		if err != nil {
			fmt.Printf("Error retrieving stats: %v\n", err)
			return
		}
		fmt.Printf("Your current stats: %v\n", stats)
	},
}

// exitCmd exits the application
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
	rootCmd.AddCommand(quizCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(scoreCmd)
	rootCmd.AddCommand(exitCmd)
}
