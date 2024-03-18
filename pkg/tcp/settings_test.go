package tcp

import (
	"errors"
	"testing"

	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/ss-server/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptrTo[T any](x T) *T { return &x }

func Test_Settings_SetDefaults(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		initial  Settings
		expected Settings
	}{
		"empty settings": {
			expected: Settings{
				Address:      ptrTo(":8388"),
				LogAddresses: ptrTo(false),
				CipherName:   core.Chacha20IetfPoly1305,
				Password:     ptrTo(""),
			},
		},
		"already set settings": {
			initial: Settings{
				Address:      ptrTo(":0"),
				LogAddresses: ptrTo(true),
				CipherName:   core.AES128gcm,
				Password:     ptrTo("password"),
			},
			expected: Settings{
				Address:      ptrTo(":0"),
				LogAddresses: ptrTo(true),
				CipherName:   core.AES128gcm,
				Password:     ptrTo("password"),
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			settings := testCase.initial

			settings.SetDefaults()

			assert.Equal(t, testCase.expected, settings)
		})
	}
}

func Test_Settings_Copy(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		original Settings
		copied   Settings
	}{
		"empty settings": {},
		"non empty settings": {
			original: Settings{
				Address:      ptrTo(":0"),
				LogAddresses: ptrTo(true),
				CipherName:   core.AES128gcm,
				Password:     ptrTo("password"),
			},
			copied: Settings{
				Address:      ptrTo(":0"),
				LogAddresses: ptrTo(true),
				CipherName:   core.AES128gcm,
				Password:     ptrTo("password"),
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			settings := testCase.original

			copied := settings.Copy()

			assert.Equal(t, testCase.copied, copied)

			// Check pointers are deep copied
			if copied.LogAddresses != nil {
				*copied.LogAddresses = !*copied.LogAddresses
				assert.NotEqual(t, copied.LogAddresses, settings.LogAddresses)
			}
			if copied.Password != nil {
				*copied.Password += "x"
				assert.NotEqual(t, copied.Password, settings.Password)
			}
		})
	}
}

func Test_Settings_OverrideWith(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		original  Settings
		other     Settings
		overidden Settings
	}{
		"empty settings with empty other": {},
		"settings with empty other": {
			original: Settings{
				Address:      ptrTo(":0"),
				LogAddresses: ptrTo(true),
				CipherName:   core.AES128gcm,
				Password:     ptrTo("password"),
			},
			overidden: Settings{
				Address:      ptrTo(":0"),
				LogAddresses: ptrTo(true),
				CipherName:   core.AES128gcm,
				Password:     ptrTo("password"),
			},
		},
		"settings with other": {
			original: Settings{
				Address:      ptrTo(":0"),
				LogAddresses: ptrTo(true),
				CipherName:   core.AES128gcm,
				Password:     ptrTo("password"),
			},
			other: Settings{
				Address:      ptrTo(":1"),
				LogAddresses: ptrTo(false),
				CipherName:   core.AES256gcm,
				Password:     ptrTo("password2"),
			},
			overidden: Settings{
				Address:      ptrTo(":1"),
				LogAddresses: ptrTo(false),
				CipherName:   core.AES256gcm,
				Password:     ptrTo("password2"),
			},
		},
		"empty settings with other": {
			other: Settings{
				Address:      ptrTo(":1"),
				LogAddresses: ptrTo(false),
				CipherName:   core.AES256gcm,
				Password:     ptrTo("password2"),
			},
			overidden: Settings{
				Address:      ptrTo(":1"),
				LogAddresses: ptrTo(false),
				CipherName:   core.AES256gcm,
				Password:     ptrTo("password2"),
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			settings := testCase.original

			settings.OverrideWith(testCase.other)

			assert.Equal(t, testCase.overidden, settings)
		})
	}
}

func Test_Settings_Validate(t *testing.T) {
	t.Parallel()

	errNothingWrapped := errors.New("")

	testCases := map[string]struct {
		settings   Settings
		errWrapped error
		errMessage string
	}{
		"invalid address": {
			settings: Settings{
				Address: ptrTo("x"),
			},
			errWrapped: errNothingWrapped,
			errMessage: "listening address: splitting host and port: " +
				"address x: missing port in address",
		},
		"invalid port": {
			settings: Settings{
				Address: ptrTo(":100000"),
			},
			errWrapped: validate.ErrPortTooHigh,
			errMessage: "listening address: port cannot be higher than 65535: 100000",
		},
		"invalid cipher": {
			settings: Settings{
				Address:    ptrTo(":0"),
				CipherName: "garbage",
			},
			errWrapped: validate.ErrValueNotOneOf,
			errMessage: "cipher: value is not one of the possible choices: " +
				"garbage must be one of aes-128-gcm, aes-256-gcm or chacha20-ietf-poly1305",
		},
		"valid settings": {
			settings: Settings{
				Address:    ptrTo(":0"),
				CipherName: core.AES128gcm,
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			settings := testCase.settings

			err := settings.Validate()

			if !errors.Is(testCase.errWrapped, errNothingWrapped) {
				require.ErrorIs(t, err, testCase.errWrapped)
			}
			if testCase.errWrapped != nil {
				require.EqualError(t, err, testCase.errMessage)
			}
		})
	}
}
