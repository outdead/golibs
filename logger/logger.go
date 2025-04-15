package logger

import (
	"fmt"
	"io"
	"time"

	"github.com/outdead/golibs/files"
	"github.com/sirupsen/logrus"
)

// Config represents the configuration structure for the logger.
// It includes settings for log level and file output configuration.
type Config struct {
	Level string `json:"level" yaml:"level"`
	File  struct {
		Path   string `json:"path"   yaml:"path"`
		Layout string `json:"layout" yaml:"layout"`
	} `json:"file" yaml:"file"`
}

// Logger wraps logrus.Logger with additional configuration and methods.
type Logger struct {
	config Config
	*logrus.Logger
}

// New creates and returns a new Logger instance with default JSON formatter.
// The returned logger has no output set by default (uses stderr).
func New() *Logger {
	logger := &Logger{
		Logger: logrus.New(),
	}
	logger.Formatter = new(logrus.JSONFormatter)

	return logger
}

// AddOutput adds additional output writer to the logger.
// This allows writing logs to multiple destinations simultaneously.
// The new writer will be used in addition to any existing outputs.
func (logger *Logger) AddOutput(w io.Writer) {
	logger.Out = io.MultiWriter(logger.Out, w)
}

// Customize configures the logger based on its configuration.
// If file output is configured, it creates and opens the log file
// using the specified path and layout for rotation.
// Returns an error if file creation fails.
func (logger *Logger) Customize() error {
	if logger.config.File.Layout != "" {
		file, err := files.CreateAndOpenFile(logger.config.File.Path, time.Now().Format(logger.config.File.Layout))
		if err != nil {
			return fmt.Errorf("create logger file hook: %w", err)
		}

		logger.AddOutput(file)
	}

	return nil
}

// Writer returns the current writer used by the logger.
// This can be used to redirect the logger's output or integrate with other systems.
func (logger *Logger) Writer() io.Writer {
	return logger.Logger.Writer()
}

// Close implements the io.Closer interface for the Logger.
// Currently, it doesn't perform any cleanup but is provided for future compatibility.
func (logger *Logger) Close() error {
	return nil
}
