package log

import (
	"bytes"
	"io"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	t.Parallel()
	level := ErrorLevel
	buffer := bytes.NewBuffer(nil)
	logger := New(level, buffer)
	assert.Equal(t, level, logger.level)
	assert.Equal(t, buffer, logger.logger.Writer())
	assert.Equal(t, log.Ldate|log.Ltime|log.Lshortfile, logger.logger.Flags())
}

func Test_SimpleLogger_log(t *testing.T) {
	t.Parallel()

	buffer := bytes.NewBuffer(nil)
	logger := &SimpleLogger{
		logger: log.New(buffer, "", 0),
	}
	logger.log(ErrorLevel, "test")

	b, err := io.ReadAll(buffer)
	require.NoError(t, err)
	written := string(b)

	const expected = "[ERROR] test\n"
	assert.Equal(t, expected, written)
}

func Test_SimpleLogger_Debug(t *testing.T) {
	t.Parallel()

	const message = "test message"

	testCases := map[string]struct {
		level    Level
		expected string
	}{
		"debug level": {
			level:    DebugLevel,
			expected: "[DEBUG] test message\n",
		},
		"info level": {
			level: InfoLevel,
		},
		"warn level": {
			level: WarnLevel,
		},
		"error level": {
			level: ErrorLevel,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			buffer := bytes.NewBuffer(nil)
			logger := &SimpleLogger{
				logger: log.New(buffer, "", 0),
				level:  testCase.level,
			}
			logger.Debug(message)

			b, err := io.ReadAll(buffer)
			require.NoError(t, err)
			written := string(b)

			assert.Equal(t, testCase.expected, written)
		})
	}
}

func Test_SimpleLogger_Info(t *testing.T) {
	t.Parallel()

	const message = "test message"

	testCases := map[string]struct {
		level    Level
		expected string
	}{
		"debug level": {
			level:    DebugLevel,
			expected: "[INFO] test message\n",
		},
		"info level": {
			level:    InfoLevel,
			expected: "[INFO] test message\n",
		},
		"warn level": {
			level: WarnLevel,
		},
		"error level": {
			level: ErrorLevel,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			buffer := bytes.NewBuffer(nil)
			logger := &SimpleLogger{
				logger: log.New(buffer, "", 0),
				level:  testCase.level,
			}
			logger.Info(message)

			b, err := io.ReadAll(buffer)
			require.NoError(t, err)
			written := string(b)

			assert.Equal(t, testCase.expected, written)
		})
	}
}

func Test_SimpleLogger_Warn(t *testing.T) {
	t.Parallel()

	const message = "test message"

	testCases := map[string]struct {
		level    Level
		expected string
	}{
		"debug level": {
			level:    DebugLevel,
			expected: "[WARN] test message\n",
		},
		"info level": {
			level:    InfoLevel,
			expected: "[WARN] test message\n",
		},
		"warn level": {
			level:    WarnLevel,
			expected: "[WARN] test message\n",
		},
		"error level": {
			level: ErrorLevel,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			buffer := bytes.NewBuffer(nil)
			logger := &SimpleLogger{
				logger: log.New(buffer, "", 0),
				level:  testCase.level,
			}
			logger.Warn(message)

			b, err := io.ReadAll(buffer)
			require.NoError(t, err)
			written := string(b)

			assert.Equal(t, testCase.expected, written)
		})
	}
}

func Test_SimpleLogger_Error(t *testing.T) {
	t.Parallel()

	const message = "test message"

	testCases := map[string]struct {
		level    Level
		expected string
	}{
		"debug level": {
			level:    DebugLevel,
			expected: "[ERROR] test message\n",
		},
		"info level": {
			level:    InfoLevel,
			expected: "[ERROR] test message\n",
		},
		"warn level": {
			level:    WarnLevel,
			expected: "[ERROR] test message\n",
		},
		"error level": {
			level:    ErrorLevel,
			expected: "[ERROR] test message\n",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			buffer := bytes.NewBuffer(nil)
			logger := &SimpleLogger{
				logger: log.New(buffer, "", 0),
				level:  testCase.level,
			}
			logger.Error(message)

			b, err := io.ReadAll(buffer)
			require.NoError(t, err)
			written := string(b)

			assert.Equal(t, testCase.expected, written)
		})
	}
}
