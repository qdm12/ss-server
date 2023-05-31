package shadowaead

import "crypto/cipher"

// SaltFilter is used to mitigate replay attacks by detecting repeated salts.
type SaltFilter interface {
	AddSalt(b []byte)
	IsSaltRepeated(b []byte) bool
}

type aeadCipher interface {
	GetSaltSize() int
	Crypt(salt []byte) (cipher.AEAD, error)
}
