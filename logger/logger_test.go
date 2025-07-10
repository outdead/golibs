package logger

import (
	"bytes"
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("should create logger with JSON formatter", func(t *testing.T) {
		log := New()

		assert.IsType(t, &logrus.JSONFormatter{}, log.Formatter)
		assert.Equal(t, logrus.InfoLevel, log.Level)
	})
}

func TestAddOutput(t *testing.T) {
	t.Run("should add multiple writers", func(t *testing.T) {
		log := New()
		log.Out = io.Discard

		buf1 := &bytes.Buffer{}
		buf2 := &bytes.Buffer{}

		log.AddOutput(buf1)
		log.AddOutput(buf2)

		log.Info("test message")

		assert.Contains(t, buf1.String(), "test message")
		assert.Contains(t, buf2.String(), "test message")
	})
}

func TestClose(t *testing.T) {
	t.Run("should not return error", func(t *testing.T) {
		log := New()
		err := log.Close()
		assert.NoError(t, err)
	})
}
