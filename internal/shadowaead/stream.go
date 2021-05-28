package shadowaead

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"net"

	"github.com/qdm12/ss-server/internal/filter"
)

const payloadSizeMask = 0x3FFF // 16KB - 1 maximum size in bytes of payload

type writer struct {
	ioWriter io.Writer
	cipher   cipher.AEAD
	buffer   []byte
	nonce    []byte
}

func newWriter(ioWriter io.Writer, aeadCipher cipher.AEAD) *writer {
	return &writer{
		ioWriter: ioWriter,
		cipher:   aeadCipher,
		buffer:   make([]byte, 2+aeadCipher.Overhead()+payloadSizeMask+aeadCipher.Overhead()),
		nonce:    make([]byte, aeadCipher.NonceSize()),
	}
}

// Write encrypts b and writes to the writer w.
func (w *writer) Write(b []byte) (int, error) {
	n, err := w.ReadFrom(bytes.NewBuffer(b))
	return int(n), err
}

// ReadFrom reads from the given reader until EOF or an error occurs, encrypts and
// writes to the writer w.
func (w *writer) ReadFrom(reader io.Reader) (n int64, err error) {
	cipherOverhead := w.cipher.Overhead()
	for {
		buf := w.buffer
		payloadBuf := buf[2+cipherOverhead : 2+cipherOverhead+payloadSizeMask]
		nr, er := reader.Read(payloadBuf)

		if nr > 0 {
			n += int64(nr)
			buf = buf[:2+cipherOverhead+nr+cipherOverhead]
			payloadBuf = payloadBuf[:nr]
			// big-endian payload size
			buf[0], buf[1] = byte(nr>>8), byte(nr) //nolint:gomnd
			w.cipher.Seal(buf[:0], w.nonce, buf[:2], nil)
			increment(w.nonce)
			w.cipher.Seal(payloadBuf[:0], w.nonce, payloadBuf, nil)
			increment(w.nonce)
			_, ew := w.ioWriter.Write(buf)
			if ew != nil {
				err = ew
				break
			}
		}

		if er != nil {
			if er != io.EOF { // ignore EOF as per io.ReaderFrom contract
				err = er
			}
			break
		}
	}

	return n, err
}

type reader struct {
	reader   io.Reader
	cipher   cipher.AEAD
	buffer   []byte
	nonce    []byte
	leftOver []byte
}

func newReader(ioReader io.Reader, cipher cipher.AEAD) *reader {
	return &reader{
		reader: ioReader,
		cipher: cipher,
		buffer: make([]byte, cipher.Overhead()+payloadSizeMask),
		nonce:  make([]byte, cipher.NonceSize()),
	}
}

func (r *reader) read() (bytesRead int, err error) {
	cipherOverhead := r.cipher.Overhead()

	// decrypt payload size
	buf := r.buffer[:2+cipherOverhead]
	if _, err := io.ReadFull(r.reader, buf); err != nil {
		return 0, err
	}

	_, err = r.cipher.Open(buf[:0], r.nonce, buf, nil)
	increment(r.nonce)
	if err != nil {
		return 0, err
	}

	size := (int(buf[0])<<8 + int(buf[1])) & payloadSizeMask

	// decrypt payload
	buf = r.buffer[:size+cipherOverhead]
	if _, err := io.ReadFull(r.reader, buf); err != nil {
		return 0, err
	}
	_, err = r.cipher.Open(buf[:0], r.nonce, buf, nil)
	increment(r.nonce)
	if err != nil {
		return 0, err
	}

	return size, nil
}

// Read reads from the reader, decrypts and writes to b.
func (r *reader) Read(b []byte) (int, error) {
	// copy decrypted bytes (if any) from previous record first
	if len(r.leftOver) > 0 {
		n := copy(b, r.leftOver)
		r.leftOver = r.leftOver[n:]
		return n, nil
	}

	n, err := r.read()
	m := copy(b, r.buffer[:n])
	if m < n { // insufficient len(b), keep leftover for next read
		r.leftOver = r.buffer[m:n]
	}
	return m, err
}

// WriteTo reads from the reader, decrypts and writes to writer until
// there is no more data to write or an error occurs.
func (r *reader) WriteTo(writer io.Writer) (n int64, err error) {
	// write decrypted bytes left over from previous record
	for len(r.leftOver) > 0 {
		bytesWritten, err := writer.Write(r.leftOver)
		r.leftOver = r.leftOver[bytesWritten:]
		n += int64(bytesWritten)
		if err != nil {
			return n, err
		}
	}
	for {
		bytesRead, readError := r.read()
		if bytesRead > 0 {
			bytesWritten, err := writer.Write(r.buffer[:bytesRead])
			n += int64(bytesWritten)
			if err != nil {
				return n, err
			}
		}
		if readError != nil {
			if readError == io.EOF {
				return n, nil // ignore EOF error
			}
			return n, readError
		}
	}
}

// increment little-endian encoded unsigned integer b and wrap around on overflow.
func increment(b []byte) {
	for i := range b {
		b[i]++
		if b[i] != 0 {
			return
		}
	}
}

// NewConn wraps a stream net.Conn connection with a cipher.
func NewConn(connection net.Conn, aead AEADCipher, saltFilter filter.SaltFilter) net.Conn {
	return &streamConn{
		Conn:       connection,
		aead:       aead,
		saltFilter: saltFilter,
	}
}

type streamConn struct {
	net.Conn
	aead       AEADCipher
	saltFilter filter.SaltFilter
	reader     *reader
	writer     *writer
}

func (c *streamConn) initReader() error {
	salt := make([]byte, c.aead.SaltSize())
	if _, err := io.ReadFull(c.Conn, salt); err != nil {
		return err
	}
	if c.saltFilter.IsSaltRepeated(salt) {
		return fmt.Errorf("possible replay attack, dropping the packet (repeated salt detected)")
	}
	aead, err := c.aead.Crypter(salt)
	if err != nil {
		return err
	}
	c.saltFilter.AddSalt(salt)

	c.reader = newReader(c.Conn, aead)
	return nil
}

func (c *streamConn) Read(b []byte) (int, error) {
	if c.reader == nil {
		if err := c.initReader(); err != nil {
			return 0, err
		}
	}
	return c.reader.Read(b)
}

func (c *streamConn) WriteTo(w io.Writer) (int64, error) {
	if c.reader == nil {
		if err := c.initReader(); err != nil {
			return 0, err
		}
	}
	return c.reader.WriteTo(w)
}

func (c *streamConn) initWriter() error {
	salt := make([]byte, c.aead.SaltSize())
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}
	aead, err := c.aead.Crypter(salt)
	if err != nil {
		return err
	}
	_, err = c.Conn.Write(salt)
	if err != nil {
		return err
	}
	c.saltFilter.AddSalt(salt)
	c.writer = newWriter(c.Conn, aead)
	return nil
}

func (c *streamConn) Write(data []byte) (int, error) {
	if c.writer == nil {
		if err := c.initWriter(); err != nil {
			return 0, err
		}
	}
	return c.writer.Write(data)
}

func (c *streamConn) ReadFrom(reader io.Reader) (int64, error) {
	if c.writer == nil {
		if err := c.initWriter(); err != nil {
			return 0, err
		}
	}
	return c.writer.ReadFrom(reader)
}
