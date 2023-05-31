package shadowaead

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/qdm12/ss-server/internal/filter"
)

//nolint:gochecknoglobals
var zeroNonce [128]byte // read-only zerored array

type cipherPacketConn struct {
	net.PacketConn
	aead       AEADCipher
	saltFilter filter.SaltFilter
	mu         sync.Mutex
	buffer     []byte // write lock
}

// NewPacketConn wraps a net.PacketConn with a cipher.
func NewPacketConn(connection net.PacketConn, aead AEADCipher, saltFilter filter.SaltFilter) net.PacketConn {
	const maxUDPPacketSize = 64 * 1024
	return &cipherPacketConn{
		PacketConn: connection,
		aead:       aead,
		buffer:     make([]byte, maxUDPPacketSize),
		saltFilter: saltFilter,
	}
}

// WriteTo encrypts b and write to addr using the embedded PacketConn.
func (c *cipherPacketConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	buf, err := c.pack(c.buffer, b)
	if err != nil {
		return 0, err
	}
	_, err = c.PacketConn.WriteTo(buf, addr)
	return len(b), err
}

// ReadFrom reads from the embedded PacketConn and decrypts into b.
func (c *cipherPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, address, err := c.PacketConn.ReadFrom(b)
	if err != nil {
		return n, address, err
	}
	bb, err := c.unpack(b[c.aead.GetSaltSize():], b[:n])
	if err != nil {
		return n, address, err
	}
	copy(b, bb)
	return len(bb), address, err
}

// pack encrypts a plaintext using the cipher provided, with a randomly generated salt and
// returns a slice of dst containing the encrypted packet.
func (c *cipherPacketConn) pack(dst, plaintext []byte) ([]byte, error) {
	saltSize := c.aead.GetSaltSize()
	salt := dst[:saltSize]
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	aead, err := c.aead.Crypt(salt)
	if err != nil {
		return nil, err
	}
	c.saltFilter.AddSalt(salt)

	if len(dst) < saltSize+len(plaintext)+aead.Overhead() {
		return nil, io.ErrShortBuffer
	}
	b := aead.Seal(dst[saltSize:saltSize], zeroNonce[:aead.NonceSize()], plaintext, nil)
	return dst[:saltSize+len(b)], nil
}

var (
	errPacketTooShort = errors.New("packet is too short")
	errRepeatedSalt   = errors.New("repeated salt detected")
)

// unpack decrypts a packet using the cipher provided and returns a slice of dst containing
// the decrypted packet.
func (c *cipherPacketConn) unpack(dst, packet []byte) (plaintext []byte, err error) {
	saltSize := c.aead.GetSaltSize()
	if len(packet) < saltSize {
		return nil, fmt.Errorf("%w: %d bytes instead of minimum of %d bytes",
			errPacketTooShort, len(packet), saltSize)
	}
	salt := packet[:saltSize]
	if c.saltFilter.IsSaltRepeated(salt) {
		return nil, fmt.Errorf("%w: possible replay attack, dropping the packet", errRepeatedSalt)
	}
	aead, err := c.aead.Crypt(salt)
	if err != nil {
		return nil, err
	}
	c.saltFilter.AddSalt(salt)
	if len(packet) < saltSize+aead.Overhead() {
		return nil, fmt.Errorf("%w: %d bytes is too short to be a valid encrypted packet",
			errPacketTooShort, len(packet))
	}
	if saltSize+len(dst)+aead.Overhead() < len(packet) {
		return nil, io.ErrShortBuffer
	}
	return aead.Open(dst[:0], zeroNonce[:aead.NonceSize()], packet[saltSize:], nil)
}
