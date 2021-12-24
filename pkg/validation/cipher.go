package validation

import (
	"errors"
	"fmt"

	"github.com/qdm12/ss-server/internal/core"
)

var (
	ErrCipherNotValid = errors.New("cipher is not valid")
)

func ValidateCipher(cipher string) (err error) {
	switch cipher {
	case core.AES128gcm, core.AES256gcm, core.Chacha20IetfPoly1305:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrCipherNotValid, cipher)
	}
}
