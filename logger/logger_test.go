package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const TestFilesDir = "./.tmp"

func setupTest(t *testing.T) func(t *testing.T) {
	os.RemoveAll(TestFilesDir)
	os.Mkdir(TestFilesDir, 0o777)

	return func(t *testing.T) {
		os.RemoveAll(TestFilesDir)
	}
}

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

func TestLogger_SetConfig(t *testing.T) {
	down := setupTest(t)
	defer down(t)

	type args struct {
		cfg     *Config
		options []Option
	}

	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "nil config error",
			args: args{
				cfg: nil,
			},
			wantErr: assert.Error,
		},
		{
			name: "config level error",
			args: args{
				cfg: &Config{
					Level: "my spoon is too big",
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "config level success",
			args: args{
				cfg: &Config{
					Level: "debug",
				},
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := New()

			tt.wantErr(t, log.SetConfig(tt.args.cfg, tt.args.options...), fmt.Sprintf("SetConfig(%v, %v)", tt.args.cfg, tt.args.options))
		})
	}

	t.Run("config file success", func(t *testing.T) {
		dir, err := os.MkdirTemp(TestFilesDir, "testdir")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		cfg := &Config{
			Level: "debug",
		}
		cfg.File.Path = dir
		cfg.File.Layout = "test.json"
		//cfg.File.Layout = DefaultFileHookLayout

		log := New()
		log.SetOutput(io.Discard)

		if err := log.SetConfig(cfg); err != nil {
			require.NoError(t, err)

			return
		}

		log.Info("test message")

		b, err := os.ReadFile(dir + "/" + cfg.File.Layout)
		if err != nil {
			require.NoError(t, err)

			return
		}

		fileData := string(b)

		require.Contains(t, fileData, "test message")
	})
}

func TestClose(t *testing.T) {
	t.Run("should not return error", func(t *testing.T) {
		log := New()
		err := log.Close()
		assert.NoError(t, err)
	})
}
