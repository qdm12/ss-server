package env

import (
	"testing"

	"github.com/qdm12/ss-server/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewReader(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		environ        []string
		expectedReader *reader
	}{
		"empty environ": {
			expectedReader: &reader{
				envKV: map[string]string{},
			},
		},
		"two elements environ": {
			environ: []string{"k1=v1", "k2=v2"},
			expectedReader: &reader{
				envKV: map[string]string{"k1": "v1", "k2": "v2"},
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			Reader := NewReader(testCase.environ)
			reader, ok := Reader.(*reader)
			require.True(t, ok)
			assert.Equal(t, testCase.expectedReader, reader)
		})
	}
}

func Test_reader_CipherName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		reader     *reader
		cipherName string
	}{
		"default": {
			reader:     &reader{},
			cipherName: "chacha20-ietf-poly1305",
		},
		"set value": {
			reader: &reader{
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
		reader   *reader
		password string
	}{
		"default": {
			reader:   &reader{},
			password: "password",
		},
		"set value": {
			reader: &reader{
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
		reader *reader
		port   string
	}{
		"default": {
			reader: &reader{},
			port:   "8388",
		},
		"set value": {
			reader: &reader{
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
		reader   *reader
		logLevel log.Level
	}{
		"default": {
			reader:   &reader{},
			logLevel: log.InfoLevel,
		},
		"set value": {
			reader: &reader{
				envKV: map[string]string{"LOG_LEVEL": "value"},
			},
			logLevel: log.Level("value"),
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			logLevel := testCase.reader.LogLevel()
			assert.Equal(t, testCase.logLevel, logLevel)
		})
	}
}

func Test_reader_Profiling(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		reader    *reader
		profiling bool
	}{
		"default": {
			reader: &reader{},
		},
		"not on": {
			reader: &reader{
				envKV: map[string]string{"PROFILING": "off"},
			},
		},
		"on": {
			reader: &reader{
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
