package config

import (
	"os"
)

type Config struct {
	Environment       string
	LogFilePath       string
	APIBaseURL        string
	ServerPort        string
	SessionSecret     string
	QuestionsFilePath string
}

func LoadConfig() Config {
	return Config{
		Environment:       getEnv("ENV", "development"),
		LogFilePath:       getEnv("LOG_FILE_PATH", "logs/app.log"),
		APIBaseURL:        getEnv("API_BASE_URL", "http://localhost:8080"),
		ServerPort:        getEnv("SERVER_PORT", ":8080"),
		SessionSecret:     getEnv("SESSION_SECRET", "quiz-secret"),
		QuestionsFilePath: getEnv("QUESTIONS_FILE_PATH", "questions.csv"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
