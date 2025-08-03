package logger

import (
	"fmt"
	"io"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/outdead/discordbotrus"
	"github.com/outdead/golibs/files"
	"github.com/sirupsen/logrus"
)

// Hook includes logrus.Hook interface and describes Close method.
type Hook interface {
	logrus.Hook
	Close() error
}

// Logger wraps logrus.Logger with additional configuration and methods.
type Logger struct {
	*logrus.Logger
	config Config

	discordHook    Hook
	discordSession *discordgo.Session
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
func (log *Logger) AddOutput(w io.Writer) {
	log.Out = io.MultiWriter(log.Out, w)
}

// SetConfig applies a new configuration to the Logger and configures all required outputs.
// It handles the following configuration aspects:
//   - Core logger settings (log level)
//   - File output configuration (if specified)
//   - Discord hook setup (if configured)
//
// Parameters:
//   - cfg: Pointer to Config struct containing all logger settings. Must not be nil.
//   - options: Optional variadic list of Option functions to modify logger behavior.
//
// Returns:
//   - error: Returns ErrInvalidConfig if cfg is nil or contains invalid settings.
//     Returns file creation errors if file logging is configured.
//     Returns Discord hook initialization errors if Discord logging is configured.
//
// Usage:
//
//	err := logger.SetConfig(&Config{
//	    Level: "info",
//	    File: FileConfig{Path: "/var/log", Layout: "2006-01-02.log"},
//	})
//	if err != nil {
//	    // handle error
//	}
//
// Notes:
//   - This function is not concurrent-safe and should not be called while the logger is in use.
//   - Replaces all previous configuration when called.
//   - File outputs are created immediately if specified in config.
//   - Discord hooks are initialized immediately if configured.
func (log *Logger) SetConfig(cfg *Config, options ...Option) error {
	if cfg == nil {
		return ErrInvalidConfig
	}

	log.config = *cfg

	for _, option := range options {
		option(log)
	}

	if cfg.Level != "" {
		logrusLevel, err := logrus.ParseLevel(cfg.Level)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidConfig, err)
		}

		log.Level = logrusLevel
	}

	if log.config.File.Layout != "" {
		file, err := files.CreateAndOpenFile(log.config.File.Path, time.Now().Format(log.config.File.Layout))
		if err != nil {
			return fmt.Errorf("create logger file hook: %w", err)
		}

		log.AddOutput(file)
	}

	if cfg.Discord.ChannelID != "" {
		var err error

		if cfg.Discord.Token != "" {
			log.discordHook, err = discordbotrus.New(&cfg.Discord)
		} else {
			log.discordHook, err = discordbotrus.New(&cfg.Discord, discordbotrus.WithSession(log.discordSession))
		}

		if err != nil {
			return fmt.Errorf("create logrus discord hook error: %w", err)
		}

		log.AddHook(log.discordHook)
	}

	return nil
}

// Writer returns the current writer used by the logger.
// This can be used to redirect the logger's output or integrate with other systems.
func (log *Logger) Writer() io.Writer {
	return log.Logger.Writer()
}

// Close implements the io.Closer interface for the Logger.
// Currently, it doesn't perform any cleanup but is provided for future compatibility.
func (log *Logger) Close() error {
	if log.discordHook != nil {
		return log.discordHook.Close()
	}

	return nil
}
