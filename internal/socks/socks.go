package socks

import (
	"fmt"
	"io"
	"net"
	"strconv"
)

// SOCKS address types
const (
	addressTypeIPv4       = 1
	addressTypeDomainName = 3
	addressTypeIPv6       = 4
)

const maxSocksAddrressLength = 1 + 1 + 255 + 2

// Address is a SOCKS address
type Address []byte

// String serializes a SOCKS address to a string
func (a Address) String() string {
	var host, port string
	switch a[0] { // address type
	case addressTypeDomainName:
		host = string(a[2 : 2+int(a[1])])
		port = strconv.Itoa((int(a[2+int(a[1])]) << 8) | int(a[2+int(a[1])+1]))
	case addressTypeIPv4:
		host = net.IP(a[1 : 1+net.IPv4len]).String()
		port = strconv.Itoa((int(a[1+net.IPv4len]) << 8) | int(a[1+net.IPv4len+1]))
	case addressTypeIPv6:
		host = net.IP(a[1 : 1+net.IPv6len]).String()
		port = strconv.Itoa((int(a[1+net.IPv6len]) << 8) | int(a[1+net.IPv6len+1]))
	}
	return net.JoinHostPort(host, port)
}

func readAddress(reader io.Reader, buffer []byte) (socksAddress Address, err error) {
	if len(buffer) < maxSocksAddrressLength {
		return nil, io.ErrShortBuffer
	}
	_, err = io.ReadFull(reader, buffer[:1]) // read 1st byte for address type
	if err != nil {
		return nil, err
	}
	switch buffer[0] {
	case addressTypeDomainName:
		_, err = io.ReadFull(reader, buffer[1:2]) // read 2nd byte for domain length
		if err != nil {
			return nil, err
		}
		_, err = io.ReadFull(reader, buffer[2:2+int(buffer[1])+2])
		return buffer[:1+1+int(buffer[1])+2], err
	case addressTypeIPv4:
		_, err = io.ReadFull(reader, buffer[1:1+net.IPv4len+2])
		return buffer[:1+net.IPv4len+2], err
	case addressTypeIPv6:
		_, err = io.ReadFull(reader, buffer[1:1+net.IPv6len+2])
		return buffer[:1+net.IPv6len+2], err
	}
	return nil, fmt.Errorf("Socks address type %b is not supported", buffer[0])
}

// ReadAddress reads bytes from the reader to get a Socks address
func ReadAddress(reader io.Reader) (socksAddress Address, err error) {
	return readAddress(reader, make([]byte, maxSocksAddrressLength))
}

// ExtractAddress extracts a SOCKS address from the beginning of a packet
func ExtractAddress(packet []byte) (socksAddress Address, err error) {
	if len(packet) == 0 {
		return nil, fmt.Errorf("cannot extract SOCKS address from empty packet")
	}
	var length int
	switch packet[0] {
	case addressTypeDomainName:
		if len(packet) <= 1 {
			return nil, fmt.Errorf("cannot extract SOCKS address from packet with 0/1 byte for a domain name type address")
		}
		length = 1 + 1 + int(packet[1]) + 2
	case addressTypeIPv4:
		length = 1 + net.IPv4len + 2
	case addressTypeIPv6:
		length = 1 + net.IPv6len + 2
	default:
		return nil, fmt.Errorf("unknown SOCKS address type %b", packet[0])
	}
	if len(packet) < length {
		return nil, fmt.Errorf("cannot extract %d bytes SOCKS address from packet with %d bytes", length, len(packet))
	}
	return packet[:length], nil
}

// ParseAddress parses the SOCKS address from a network address
func ParseAddress(remoteAddress net.Addr) (socksAddress Address, err error) {
	s := remoteAddress.String()
	host, portStr, err := net.SplitHostPort(s)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(host)
	var ipv4 net.IP
	if ip != nil {
		ipv4 = ip.To4()
	}
	switch {
	case ipv4 != nil:
		socksAddress = make([]byte, 1+net.IPv4len+2)
		socksAddress[0] = addressTypeIPv4
		copy(socksAddress[1:], ipv4)
	case ip != nil: // ipv6
		socksAddress = make([]byte, 1+net.IPv6len+2)
		socksAddress[0] = addressTypeIPv6
		copy(socksAddress[1:], ip)
	case len(host) > 255:
		return nil, fmt.Errorf("parsed host %q cannot be longer than 255 characters", host)
	default:
		socksAddress = make([]byte, 1+1+len(host)+2)
		socksAddress[0] = addressTypeDomainName
		socksAddress[1] = byte(len(host))
		copy(socksAddress[2:], host)
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("cannot parse port %q: %w", portStr, err)
	}
	socksAddress[len(socksAddress)-2] = byte(port >> 8)
	socksAddress[len(socksAddress)-1] = byte(port)
	return socksAddress, nil
}

// Handshake fast-tracks SOCKS initialization to get target address to connect.
func Handshake(readWriter io.ReadWriter) (targetAddress Address, err error) {
	// Read RFC 1928 for request and reply structure and sizes.
	buffer := make([]byte, maxSocksAddrressLength)
	// read VER, NMETHODS, METHODS
	if _, err := io.ReadFull(readWriter, buffer[:2]); err != nil {
		return nil, fmt.Errorf("cannot handshake: %w", err)
	}
	nmethods := buffer[1]
	if _, err := io.ReadFull(readWriter, buffer[:nmethods]); err != nil {
		return nil, fmt.Errorf("cannot handshake: %w", err)
	}
	// write VER METHOD
	if _, err := readWriter.Write([]byte{5, 0}); err != nil {
		return nil, fmt.Errorf("cannot handshake: %w", err)
	}
	// read VER CMD RSV ATYP DST.ADDR DST.PORT
	if _, err := io.ReadFull(readWriter, buffer[:3]); err != nil {
		return nil, fmt.Errorf("cannot handshake: %w", err)
	}
	targetAddress, err = readAddress(readWriter, buffer)
	if err != nil {
		return nil, fmt.Errorf("cannot handshake: %w", err)
	}

	const (
		CommandConnect      = 1
		CommandBind         = 2 // not supported
		CommandUDPAssociate = 3 // not supported
	)
	switch buffer[1] {
	case CommandConnect:
		replySuccess := []byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0}
		_, err = readWriter.Write(replySuccess)
		if err != nil {
			return nil, fmt.Errorf("cannot handshake: %w", err)
		}
	default:
		return nil, fmt.Errorf("cannot handshake: SOCKS command %b is not supported", buffer[1])
	}
	return targetAddress, nil // skip VER, CMD, RSV fields
}
