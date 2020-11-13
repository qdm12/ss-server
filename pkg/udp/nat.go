package udp

import (
	"net"
	"time"

	"sync"

	"github.com/qdm12/ss-server/internal/log"
	"github.com/qdm12/ss-server/internal/socks"
)

// Packet NAT table
type natmap struct {
	mu                        sync.RWMutex
	remoteAddressToConnection map[string]net.PacketConn
	timeNow                   func() time.Time
}

func (nm *natmap) Get(key string) net.PacketConn {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	return nm.remoteAddressToConnection[key]
}

func (nm *natmap) Set(key string, packetConnection net.PacketConn) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.remoteAddressToConnection[key] = packetConnection
}

func (nm *natmap) Del(remoteAddress string) (packetConnection net.PacketConn) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	packetConnection = nm.remoteAddressToConnection[remoteAddress]
	delete(nm.remoteAddressToConnection, remoteAddress)
	return packetConnection // can be nil
}

func (nm *natmap) Handle(peer net.Addr, dst, src net.PacketConn, logger log.Logger) {
	_ = timedCopy(dst, peer, src, nm.timeNow)
	key := peer.String()
	nm.mu.Lock()
	packetConnection := nm.remoteAddressToConnection[key]
	delete(nm.remoteAddressToConnection, key)
	nm.mu.Unlock()
	if packetConnection != nil {
		packetConnection.Close()
	}
}

// copy from src to dst at target with read timeout
func timedCopy(dst net.PacketConn, target net.Addr, src net.PacketConn, timeNow func() time.Time) error {
	const timeout = time.Minute
	buffer := make([]byte, bufferSize)
	for {
		if err := src.SetReadDeadline(timeNow().Add(timeout)); err != nil {
			return err
		}
		bytesRead, remoteAddress, err := src.ReadFrom(buffer)
		if err != nil {
			return err
		}

		// add original packet source
		srcAddr, err := socks.ParseAddress(remoteAddress)
		if err != nil {
			return err
		}
		copy(buffer[len(srcAddr):], buffer[:bytesRead])
		copy(buffer, srcAddr)
		if _, err := dst.WriteTo(buffer[:len(srcAddr)+bytesRead], target); err != nil {
			return err
		}
	}
}
