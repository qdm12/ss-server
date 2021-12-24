package validation

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ValidateAddress(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		address    string
		errWrapped error
		errMessage string
	}{
		"missing port": {
			address:    "host",
			errWrapped: ErrListenAddressNotValid,
			errMessage: "listening address is not valid: address host: missing port in address",
		},
		"port not an integer": {
			address:    "host:a",
			errWrapped: ErrListenPortNotValid,
			errMessage: "listening port is not valid: strconv.Atoi: parsing \"a\": invalid syntax",
		},
		"port negative": {
			address:    "host:-1",
			errWrapped: ErrListenPortNotValid,
			errMessage: "listening port is not valid: -1: must be between 0 and 65535",
		},
		"port too big": {
			address:    "host:65536",
			errWrapped: ErrListenPortNotValid,
			errMessage: "listening port is not valid: 65536: must be between 0 and 65535",
		},
		"success": {
			address: "host:65535",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := ValidateAddress(testCase.address)

			assert.ErrorIs(t, err, testCase.errWrapped)
			if testCase.errWrapped != nil {
				assert.EqualError(t, err, testCase.errMessage)
			}
		})
	}

	t.Run("privileged without root", func(t *testing.T) {
		t.Parallel()

		uid := os.Getuid()
		if uid == 0 {
			t.Skip("skipping because we are running as root")
		}

		const address = "host:1000"
		err := ValidateAddress(address)
		assert.ErrorIs(t, err, ErrListenPortPrivileged)
		assert.EqualError(t, err, "cannot use a privileged listening port "+
			"without running as root: "+
			"port 1000 with user id "+fmt.Sprint(uid))
	})
}
