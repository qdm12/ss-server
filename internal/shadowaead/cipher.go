package shadowaead

//nolint:gci
import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1" //nolint: gosec
	"io"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

type AEADCipher interface {
	SaltSize() int
	Crypter(salt []byte) (cipher.AEAD, error)
}

type aeadCipherAdapter struct {
	preSharedKey  []byte
	newAEADCipher func(key []byte) (cipher.AEAD, error)
}

func (c *aeadCipherAdapter) keySize() int {
	return len(c.preSharedKey)
}

func (c *aeadCipherAdapter) SaltSize() int {
	const minimumSaltSize = 16
	if ks := c.keySize(); ks > minimumSaltSize {
		return ks
	}
	return minimumSaltSize
}

func (c *aeadCipherAdapter) Crypter(salt []byte) (cipher.AEAD, error) {
	subkey := make([]byte, c.keySize())
	const keyInfo = "ss-subkey"
	reader := hkdf.New(sha1.New, c.preSharedKey, salt, []byte(keyInfo))
	_, _ = io.ReadFull(reader, subkey)
	return c.newAEADCipher(subkey)
}

// Chacha20Poly1305 creates a new Cipher with a pre-shared key of 32 bytes.
func Chacha20Poly1305(preSharedKey []byte) AEADCipher {
	return &aeadCipherAdapter{
		preSharedKey:  preSharedKey,
		newAEADCipher: chacha20poly1305.New,
	}
}

// AESGCM creates a new Cipher with a pre-shared key of 16 or 32 bytes.
func AESGCM(preSharedKey []byte) AEADCipher {
	return &aeadCipherAdapter{
		preSharedKey:  preSharedKey,
		newAEADCipher: newAESGCM,
	}
}

func newAESGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
