package utils

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitializeLogger sets up the global logger with environment-based configuration.
func InitializeLogger(env string, logFilePath string) (*zap.Logger, error) {
	// Ensure the directory for the log file exists
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	// Set the log level based on the environment
	var level zapcore.Level
	var encoderConfig zapcore.EncoderConfig
	if env == "production" {
		level = zap.InfoLevel
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		level = zap.DebugLevel
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // Human-readable time format

	// Create the file writer
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	fileWriter := zapcore.AddSync(file)

	// Set up the core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // Use JSON encoding
		fileWriter,                            // Log to file
		level,                                 // Log level
	)

	// Add console output for development
	if env != "production" {
		consoleWriter := zapcore.Lock(os.Stdout)
		core = zapcore.NewTee(
			core,
			zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), consoleWriter, level),
		)
	}

	// Build the logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	// Replace the global logger with the configured one
	zap.ReplaceGlobals(logger)

	return logger, nil
}
