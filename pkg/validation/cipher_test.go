package validation

import (
	"testing"

	"github.com/qdm12/ss-server/internal/core"
	"github.com/stretchr/testify/assert"
)

func Test_ValidateCipher(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		cipher     string
		errWrapped error
		errMessage string
	}{
		"aes 128 gcm": {
			cipher: core.AES128gcm,
		},
		"aes 256 gcm": {
			cipher: core.AES256gcm,
		},
		"chcha20 ietf poly 1305": {
			cipher: core.Chacha20IetfPoly1305,
		},
		"not valid": {
			cipher:     "garbage",
			errWrapped: ErrCipherNotValid,
			errMessage: "cipher is not valid: garbage",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := ValidateCipher(testCase.cipher)

			assert.ErrorIs(t, err, testCase.errWrapped)
			if testCase.errWrapped != nil {
				assert.EqualError(t, err, testCase.errMessage)
			}
		})
	}
}
