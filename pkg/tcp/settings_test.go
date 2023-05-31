package tcp

import (
	"testing"

	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/govalid/address"
	"github.com/qdm12/govalid/port"
	"github.com/qdm12/ss-server/internal/core"
	"github.com/stretchr/testify/assert"
)

func boolPtr(t bool) *bool       { return &t }
func stringPtr(s string) *string { return &s }

func Test_Settings_SetDefaults(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		initial  Settings
		expected Settings
	}{
		"empty settings": {
			expected: Settings{
				Address:      ":8388",
				LogAddresses: boolPtr(false),
				CipherName:   core.Chacha20IetfPoly1305,
				Password:     stringPtr(""),
			},
		},
		"already set settings": {
			initial: Settings{
				Address:      ":0",
				LogAddresses: boolPtr(true),
				CipherName:   core.AES128gcm,
				Password:     stringPtr("password"),
			},
			expected: Settings{
				Address:      ":0",
				LogAddresses: boolPtr(true),
				CipherName:   core.AES128gcm,
				Password:     stringPtr("password"),
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
				Address:      ":0",
				LogAddresses: boolPtr(true),
				CipherName:   core.AES128gcm,
				Password:     stringPtr("password"),
			},
			copied: Settings{
				Address:      ":0",
				LogAddresses: boolPtr(true),
				CipherName:   core.AES128gcm,
				Password:     stringPtr("password"),
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

func Test_Settings_MergeWith(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		original Settings
		other    Settings
		merged   Settings
	}{
		"empty settings with empty other": {},
		"settings with empty other": {
			original: Settings{
				Address:      ":0",
				LogAddresses: boolPtr(true),
				CipherName:   core.AES128gcm,
				Password:     stringPtr("password"),
			},
			merged: Settings{
				Address:      ":0",
				LogAddresses: boolPtr(true),
				CipherName:   core.AES128gcm,
				Password:     stringPtr("password"),
			},
		},
		"settings with other": {
			original: Settings{
				Address:      ":0",
				LogAddresses: boolPtr(true),
				CipherName:   core.AES128gcm,
				Password:     stringPtr("password"),
			},
			other: Settings{
				Address:      ":1",
				LogAddresses: boolPtr(false),
				CipherName:   core.AES256gcm,
				Password:     stringPtr("password2"),
			},
			merged: Settings{
				Address:      ":0",
				LogAddresses: boolPtr(true),
				CipherName:   core.AES128gcm,
				Password:     stringPtr("password"),
			},
		},
		"empty settings with other": {
			other: Settings{
				Address:      ":1",
				LogAddresses: boolPtr(false),
				CipherName:   core.AES256gcm,
				Password:     stringPtr("password2"),
			},
			merged: Settings{
				Address:      ":1",
				LogAddresses: boolPtr(false),
				CipherName:   core.AES256gcm,
				Password:     stringPtr("password2"),
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			settings := testCase.original.Copy()

			merged := settings.MergeWith(testCase.other)

			assert.Equal(t, testCase.original, settings)
			assert.Equal(t, testCase.merged, merged)
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
				Address:      ":0",
				LogAddresses: boolPtr(true),
				CipherName:   core.AES128gcm,
				Password:     stringPtr("password"),
			},
			overidden: Settings{
				Address:      ":0",
				LogAddresses: boolPtr(true),
				CipherName:   core.AES128gcm,
				Password:     stringPtr("password"),
			},
		},
		"settings with other": {
			original: Settings{
				Address:      ":0",
				LogAddresses: boolPtr(true),
				CipherName:   core.AES128gcm,
				Password:     stringPtr("password"),
			},
			other: Settings{
				Address:      ":1",
				LogAddresses: boolPtr(false),
				CipherName:   core.AES256gcm,
				Password:     stringPtr("password2"),
			},
			overidden: Settings{
				Address:      ":1",
				LogAddresses: boolPtr(false),
				CipherName:   core.AES256gcm,
				Password:     stringPtr("password2"),
			},
		},
		"empty settings with other": {
			other: Settings{
				Address:      ":1",
				LogAddresses: boolPtr(false),
				CipherName:   core.AES256gcm,
				Password:     stringPtr("password2"),
			},
			overidden: Settings{
				Address:      ":1",
				LogAddresses: boolPtr(false),
				CipherName:   core.AES256gcm,
				Password:     stringPtr("password2"),
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

	testCases := map[string]struct {
		settings   Settings
		errWrapped error
		errMessage string
	}{
		"invalid address": {
			settings: Settings{
				Address: "",
			},
			errWrapped: address.ErrValueNotValid,
			errMessage: "listening address: value is not valid: missing port in address",
		},
		"invalid port": {
			settings: Settings{
				Address: ":100000",
			},
			errWrapped: port.ErrPortTooHigh,
			errMessage: "listening address: port cannot be higher than 65535: 100000",
		},
		"invalid cipher": {
			settings: Settings{
				Address:    ":0",
				CipherName: "garbage",
			},
			errWrapped: validate.ErrValueNotOneOf,
			errMessage: "cipher: value is not one of the possible choices: " +
				"garbage must be one of aes-128-gcm, aes-256-gcm or chacha20-ietf-poly1305",
		},
		"valid settings": {
			settings: Settings{
				Address:    ":0",
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

			assert.ErrorIs(t, err, testCase.errWrapped)
			if testCase.errWrapped != nil {
				assert.EqualError(t, err, testCase.errMessage)
			}
		})
	}
}
