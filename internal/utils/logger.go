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
	if env == "production" {
		level = zap.InfoLevel
	} else {
		level = zap.DebugLevel
	}

	// Encoder configuration for Common Log Format (CLF)
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("02/Jan/2006:15:04:05 -0700"), // CLF time format
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create the file writer
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	fileWriter := zapcore.AddSync(file)

	// Create the core with custom format
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig), // Use CLF-like console encoder
		fileWriter,                               // Log to file
		level,                                    // Log level
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
