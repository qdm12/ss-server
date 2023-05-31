package env

import (
	"testing"

	"github.com/qdm12/log"
	"github.com/stretchr/testify/assert"
)

func Test_NewReader(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		environ        []string
		expectedReader *Reader
	}{
		"empty environ": {
			expectedReader: &Reader{
				envKV: map[string]string{},
			},
		},
		"two elements environ": {
			environ: []string{"k1=v1", "k2=v2"},
			expectedReader: &Reader{
				envKV: map[string]string{"k1": "v1", "k2": "v2"},
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			reader := NewReader(testCase.environ)
			assert.Equal(t, testCase.expectedReader, reader)
		})
	}
}

func Test_reader_CipherName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		reader     *Reader
		cipherName string
	}{
		"default": {
			reader:     &Reader{},
			cipherName: "chacha20-ietf-poly1305",
		},
		"set value": {
			reader: &Reader{
				envKV: map[string]string{"CIPHER": "value"},
			},
			cipherName: "value",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cipherName := testCase.reader.CipherName()
			assert.Equal(t, testCase.cipherName, cipherName)
		})
	}
}

func Test_reader_Password(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		reader   *Reader
		password string
	}{
		"default": {
			reader:   &Reader{},
			password: "password",
		},
		"set value": {
			reader: &Reader{
				envKV: map[string]string{"PASSWORD": "value"},
			},
			password: "value",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			password := testCase.reader.Password()
			assert.Equal(t, testCase.password, password)
		})
	}
}

func Test_reader_Port(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		reader *Reader
		port   string
	}{
		"default": {
			reader: &Reader{},
			port:   "8388",
		},
		"set value": {
			reader: &Reader{
				envKV: map[string]string{"PORT": "value"},
			},
			port: "value",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			port := testCase.reader.Port()
			assert.Equal(t, testCase.port, port)
		})
	}
}

func Test_reader_LogLevel(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		reader     *Reader
		logLevel   log.Level
		errWrapped error
		errMessage string
	}{
		"default": {
			reader:   &Reader{},
			logLevel: log.LevelInfo,
		},
		"valid_value": {
			reader: &Reader{
				envKV: map[string]string{"LOG_LEVEL": "warn"},
			},
			logLevel: log.LevelWarn,
		},
		"invalid_value": {
			reader: &Reader{
				envKV: map[string]string{"LOG_LEVEL": "xxx"},
			},
			errWrapped: ErrLogLevelUnknown,
			errMessage: "log level is unknown: xxx",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			logLevel, err := testCase.reader.LogLevel()

			assert.ErrorIs(t, err, testCase.errWrapped)
			if testCase.errWrapped != nil {
				assert.EqualError(t, err, testCase.errMessage)
			}
			assert.Equal(t, testCase.logLevel, logLevel)
		})
	}
}

func Test_reader_Profiling(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		reader    *Reader
		profiling bool
	}{
		"default": {
			reader: &Reader{},
		},
		"not on": {
			reader: &Reader{
				envKV: map[string]string{"PROFILING": "off"},
			},
		},
		"on": {
			reader: &Reader{
				envKV: map[string]string{"PROFILING": "on"},
			},
			profiling: true,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			profiling := testCase.reader.Profiling()
			assert.Equal(t, testCase.profiling, profiling)
		})
	}
}
