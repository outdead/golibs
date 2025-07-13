package bash

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	NonExistentProcessPID  = "non-existent process pid 98246753798877567"
	NonExistentProcessName = "non-existent process name 98246753798877567"
)

func TestStrip(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "strip color codes",
			input:    "\033[32mHello\033[0m\n",
			expected: "Hello",
		},
		{
			name:     "no ansi codes",
			input:    "Plain text",
			expected: "Plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Strip(tt.input))
		})
	}
}

func TestGetLargeFileList(t *testing.T) {
	resp, err := GetLargeFileList(".", ".go", 10)
	assert.Nil(t, err)

	assert.Contains(t, resp, "bash.go")
}

func TestPidofByProcess(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		expectPid := strconv.Itoa(os.Getpid())
		expectName := filepath.Base(os.Args[0])

		pid, err := PidofByProcess(expectName)
		assert.Nil(t, err)
		assert.Equal(t, expectPid, pid)
	})

	t.Run("expect error", func(t *testing.T) {
		_, err := PidofByProcess(NonExistentProcessName)
		assert.NotNil(t, err)
	})
}

func TestPidofByProcessAndParam(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		expectPid := strconv.Itoa(os.Getpid())
		expectName := filepath.Base(os.Args[0])

		pid, err := PidofByProcessAndParam(expectName, expectPid)
		assert.Nil(t, err)
		assert.Equal(t, expectPid, pid)
	})

	t.Run("expect error", func(t *testing.T) {
		_, err := PidofByProcess(NonExistentProcessName)
		assert.NotNil(t, err)
	})
}

func TestCpuPercentByPID(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		expectName := filepath.Base(os.Args[0])

		pid, err := PidofByProcess(expectName)
		assert.Nil(t, err)

		_, err = CPUPercentByPID(pid)
		assert.Nil(t, err)
	})

	t.Run("expect error", func(t *testing.T) {
		_, err := CPUPercentByPID(NonExistentProcessPID)
		assert.NotNil(t, err)
	})
}

func TestGetUptimeByPID(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		expectName := filepath.Base(os.Args[0])

		pid, err := PidofByProcess(expectName)
		assert.Nil(t, err)

		_, err = GetUptimeByPID(pid)
		assert.Nil(t, err)
	})

	t.Run("expect error", func(t *testing.T) {
		_, err := GetUptimeByPID(NonExistentProcessPID)
		assert.NotNil(t, err)
	})
}

func TestMemPercentByPID(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		expectName := filepath.Base(os.Args[0])

		pid, err := PidofByProcess(expectName)
		assert.Nil(t, err)

		_, err = MemPercentByPID(pid)
		assert.Nil(t, err)
	})

	t.Run("expect error", func(t *testing.T) {
		_, err := MemPercentByPID(NonExistentProcessPID)
		assert.NotNil(t, err)
	})
}

func TestMemUsedByPID(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		expectName := filepath.Base(os.Args[0])

		pid, err := PidofByProcess(expectName)
		assert.Nil(t, err)

		_, err = MemUsedByPID(pid)
		assert.Nil(t, err)
	})

	t.Run("expect error", func(t *testing.T) {
		_, err := MemUsedByPID(NonExistentProcessPID)
		assert.NotNil(t, err)
	})
}

func TestMemUsed(t *testing.T) {
	_, err := MemUsed()
	assert.Nil(t, err)
}

func TestMemAvail(t *testing.T) {
	_, err := MemAvail()
	assert.Nil(t, err)
}
