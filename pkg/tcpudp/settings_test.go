package tcpudp

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/ss-server/internal/core"
	"github.com/qdm12/ss-server/pkg/tcp"
	"github.com/qdm12/ss-server/pkg/udp"
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
				TCP: tcp.Settings{
					Address:      ptrTo(":8388"),
					LogAddresses: ptrTo(false),
					CipherName:   core.Chacha20IetfPoly1305,
					Password:     ptrTo(""),
					Dialer:       &net.Dialer{},
				},
				UDP: udp.Settings{
					Address:      ptrTo(":8388"),
					LogAddresses: ptrTo(false),
					CipherName:   core.Chacha20IetfPoly1305,
					Password:     ptrTo(""),
				},
			},
		},
		"already set settings": {
			initial: Settings{
				Address:      ptrTo(":0"),
				LogAddresses: ptrTo(true),
				CipherName:   core.AES128gcm,
				Password:     ptrTo("password"),
				TCP: tcp.Settings{
					Address:      ptrTo(":8388"),
					LogAddresses: ptrTo(true),
					CipherName:   core.Chacha20IetfPoly1305,
					Password:     ptrTo("tcp"),
					Dialer:       &net.Dialer{Timeout: time.Second},
				},
				UDP: udp.Settings{
					Address:      ptrTo(":8388"),
					LogAddresses: ptrTo(false),
					CipherName:   core.Chacha20IetfPoly1305,
					Password:     ptrTo("udp"),
				},
			},
			expected: Settings{
				Address:      ptrTo(":0"),
				LogAddresses: ptrTo(true),
				CipherName:   core.AES128gcm,
				Password:     ptrTo("password"),
				TCP: tcp.Settings{
					Address:      ptrTo(":8388"),
					LogAddresses: ptrTo(true),
					CipherName:   core.Chacha20IetfPoly1305,
					Password:     ptrTo("tcp"),
					Dialer:       &net.Dialer{Timeout: time.Second},
				},
				UDP: udp.Settings{
					Address:      ptrTo(":8388"),
					LogAddresses: ptrTo(false),
					CipherName:   core.Chacha20IetfPoly1305,
					Password:     ptrTo("udp"),
				},
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
				TCP: tcp.Settings{
					Address:      ptrTo(":8388"),
					LogAddresses: ptrTo(true),
					CipherName:   core.Chacha20IetfPoly1305,
					Password:     ptrTo("tcp"),
				},
				UDP: udp.Settings{
					Address:      ptrTo(":8388"),
					LogAddresses: ptrTo(false),
					CipherName:   core.Chacha20IetfPoly1305,
					Password:     ptrTo("udp"),
				},
			},
			copied: Settings{
				Address:      ptrTo(":0"),
				LogAddresses: ptrTo(true),
				CipherName:   core.AES128gcm,
				Password:     ptrTo("password"),
				TCP: tcp.Settings{
					Address:      ptrTo(":8388"),
					LogAddresses: ptrTo(true),
					CipherName:   core.Chacha20IetfPoly1305,
					Password:     ptrTo("tcp"),
				},
				UDP: udp.Settings{
					Address:      ptrTo(":8388"),
					LogAddresses: ptrTo(false),
					CipherName:   core.Chacha20IetfPoly1305,
					Password:     ptrTo("udp"),
				},
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
				Address:      ptrTo(":1"),
				LogAddresses: ptrTo(true),
				CipherName:   core.AES128gcm,
				Password:     ptrTo("password"),
				TCP: tcp.Settings{
					Address: ptrTo("2"),
				},
				UDP: udp.Settings{
					Address: ptrTo("3"),
				},
			},
			overidden: Settings{
				Address:      ptrTo(":1"),
				LogAddresses: ptrTo(true),
				CipherName:   core.AES128gcm,
				Password:     ptrTo("password"),
				TCP: tcp.Settings{
					Address: ptrTo("2"),
				},
				UDP: udp.Settings{
					Address: ptrTo("3"),
				},
			},
		},
		"settings with other": {
			original: Settings{
				Address:      ptrTo(":1"),
				LogAddresses: ptrTo(true),
				CipherName:   core.AES128gcm,
				Password:     ptrTo("password"),
				TCP: tcp.Settings{
					Address: ptrTo("2"),
				},
				UDP: udp.Settings{
					Address: ptrTo("3"),
				},
			},
			other: Settings{
				Address:      ptrTo(":6"),
				LogAddresses: ptrTo(true),
				CipherName:   core.AES128gcm,
				Password:     ptrTo("password"),
				TCP: tcp.Settings{
					Address: ptrTo(":7"),
				},
				UDP: udp.Settings{
					Address: ptrTo(":8"),
				},
			},
			overidden: Settings{
				Address:      ptrTo(":6"),
				LogAddresses: ptrTo(true),
				CipherName:   core.AES128gcm,
				Password:     ptrTo("password"),
				TCP: tcp.Settings{
					Address: ptrTo(":7"),
				},
				UDP: udp.Settings{
					Address: ptrTo(":8"),
				},
			},
		},
		"empty settings with other": {
			other: Settings{
				Address:      ptrTo(":1"),
				LogAddresses: ptrTo(true),
				CipherName:   core.AES128gcm,
				Password:     ptrTo("password"),
				TCP: tcp.Settings{
					Address: ptrTo("2"),
				},
				UDP: udp.Settings{
					Address: ptrTo("3"),
				},
			},
			overidden: Settings{
				Address:      ptrTo(":1"),
				LogAddresses: ptrTo(true),
				CipherName:   core.AES128gcm,
				Password:     ptrTo("password"),
				TCP: tcp.Settings{
					Address: ptrTo("2"),
				},
				UDP: udp.Settings{
					Address: ptrTo("3"),
				},
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

	errNothingWrapped := errors.New("nothing wrapped")

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
			errMessage: "listening address: splitting host and port: address x: missing port in address",
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
		"invalid TCP": {
			settings: Settings{
				Address:    ptrTo(":0"),
				CipherName: core.AES128gcm,
				TCP: tcp.Settings{
					Address: ptrTo("garbage"),
				},
			},
			errWrapped: errNothingWrapped,
			errMessage: "TCP server settings: listening address: splitting host and port: " +
				"address garbage: missing port in address",
		},
		"invalid UDP": {
			settings: Settings{
				Address:    ptrTo(":0"),
				CipherName: core.AES128gcm,
				TCP: tcp.Settings{
					Address:    ptrTo(":0"),
					CipherName: core.AES128gcm,
				},
				UDP: udp.Settings{
					Address: ptrTo("garbage"),
				},
			},
			errWrapped: errNothingWrapped,
			errMessage: "UDP server settings: listening address: splitting host and port: " +
				"address garbage: missing port in address",
		},
		"valid settings": {
			settings: Settings{
				Address:    ptrTo(":0"),
				CipherName: core.AES128gcm,
				TCP: tcp.Settings{
					Address:    ptrTo(":0"),
					CipherName: core.AES128gcm,
				},
				UDP: udp.Settings{
					Address:    ptrTo(":0"),
					CipherName: core.AES256gcm,
				},
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
