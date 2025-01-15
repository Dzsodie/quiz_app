package utils

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/Dzsodie/quiz_app/internal/models"
	"go.uber.org/zap"
)

// Global Logger
var Logger *zap.Logger

// SetLogger sets the logger for utils.
func SetLogger(logger *zap.Logger) {
	Logger = logger
}

// ReadCSV reads questions from a CSV file and returns a slice of questions.
func ReadCSV(filename string) ([]models.Question, error) {
	if Logger == nil {
		return nil, fmt.Errorf("logger is not set")
	}

	Logger.Info("Opening CSV file", zap.String("filename", filename))
	file, err := os.Open(filename)
	if err != nil {
		Logger.Error("Failed to open file", zap.String("filename", filename), zap.Error(err))
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	Logger.Info("Reading CSV file", zap.String("filename", filename))
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		Logger.Error("Failed to read CSV file", zap.String("filename", filename), zap.Error(err))
		return nil, fmt.Errorf("failed to read CSV file: %w", err)
	}

	if len(records) == 0 {
		Logger.Warn("CSV file is empty", zap.String("filename", filename))
		return nil, fmt.Errorf("CSV file is empty")
	}

	var questions []models.Question
	for i, record := range records {
		if i == 0 { // Skip header
			continue
		}
		if len(record) < 5 {
			Logger.Warn("Invalid record in CSV file", zap.String("filename", filename), zap.Int("line", i+1), zap.Any("record", record))
			return nil, fmt.Errorf("invalid record: %v", record)
		}
		answer, err := strconv.Atoi(record[4])
		if err != nil {
			Logger.Warn("Invalid answer format in record", zap.String("filename", filename), zap.Int("line", i+1), zap.Any("record", record), zap.Error(err))
			return nil, fmt.Errorf("invalid answer format in record %v: %w", record, err)
		}
		questions = append(questions, models.Question{
			Question: record[0],
			Options:  record[1:4],
			Answer:   answer,
		})
		Logger.Debug("Processed record", zap.String("filename", filename), zap.Int("line", i+1), zap.Any("question", record[0]))
	}

	Logger.Info("CSV file processed successfully", zap.String("filename", filename), zap.Int("total_questions", len(questions)))
	return questions, nil
}
