package utils

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/Dzsodie/quiz_app/internal/models"
	"go.uber.org/zap"
)

func ReadCSV(filename string) ([]models.Question, error) {
	logger := GetLogger().Sugar()

	if logger == nil {
		return nil, fmt.Errorf("logger is not set")
	}

	logger.Info("Opening CSV file", zap.String("filename", filename))
	file, err := os.Open(filename)
	if err != nil {
		logger.Error("Failed to open file", zap.String("filename", filename), zap.Error(err))
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	logger.Info("Reading CSV file", zap.String("filename", filename))
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		logger.Error("Failed to read CSV file", zap.String("filename", filename), zap.Error(err))
		return nil, fmt.Errorf("failed to read CSV file: %w", err)
	}

	if len(records) == 0 {
		logger.Warn("CSV file is empty", zap.String("filename", filename))
		return nil, fmt.Errorf("CSV file is empty: %s", filename)
	}

	var questions []models.Question
	for i, record := range records {
		if i == 0 {
			continue
		}
		if len(record) < 6 {
			logger.Warn("Invalid record in CSV file", zap.String("filename", filename), zap.Int("line", i+1), zap.Any("record", record))
			return nil, fmt.Errorf("invalid answer format in record: %v", record)
		}
		answer, err := strconv.Atoi(record[5])
		if err != nil || answer < 1 || answer > 3 {
			logger.Warn("Invalid answer format in record", zap.String("filename", filename), zap.Int("line", i+1), zap.Any("record", record), zap.Error(err))
			return nil, fmt.Errorf("invalid answer format in record %v: %w", record, err)
		}

		questions = append(questions, models.Question{
			QuestionID: i,
			Question:   record[1],
			Options:    record[2:5],
			Answer:     answer,
		})
		logger.Debug("Processed record", zap.String("filename", filename), zap.Int("line", i+1), zap.Any("question", record[1]))
	}

	logger.Info("CSV file processed successfully", zap.String("filename", filename), zap.Int("total_questions", len(questions)))
	return questions, nil
}
