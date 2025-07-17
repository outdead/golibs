package logger

import (
	"errors"

	"github.com/outdead/discordbotrus"
)

const DefaultFileHookLayout = "20060102_log.json"

var ErrInvalidConfig = errors.New("invalid config")

// Config represents the configuration structure for the logger.
// It includes settings for log level and file output configuration.
type Config struct {
	Level string `json:"level" yaml:"level"`
	File  struct {
		Path   string `json:"path"   yaml:"path"`
		Layout string `json:"layout" yaml:"layout"`
	} `json:"file" yaml:"file"`
	Discord discordbotrus.Config `json:"discord" yaml:"discord"`
}
