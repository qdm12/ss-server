package udp

import (
	"context"
	"net"
	"time"

	"github.com/qdm12/ss-server/internal/core"
	"github.com/qdm12/ss-server/internal/filter"
	"github.com/qdm12/ss-server/internal/socks"
	"github.com/qdm12/ss-server/pkg/log"
)

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE . Listener

var _ Listener = (*Server)(nil)

type Listener interface {
	Listen(ctx context.Context, address string) (err error)
}

func NewServer(cipherName, password string, logger log.Logger) (s *Server, err error) {
	udpPacketCipher, err := core.NewUDPPacketCipher(cipherName, password, filter.NewBloomRing())
	if err != nil {
		return nil, err
	}
	return &Server{
		logger:   logger,
		timeNow:  time.Now,
		shadower: udpPacketCipher,
	}, nil
}

type Server struct {
	logger   log.Logger
	timeNow  func() time.Time
	shadower core.PacketConnShadower
}

// Listen listens on the address given for encrypted packets and does UDP NATing.
func (s *Server) Listen(ctx context.Context, address string) (err error) {
	listenConfig := net.ListenConfig{}
	packetConnection, err := listenConfig.ListenPacket(ctx, "udp", address)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		if err := packetConnection.Close(); err != nil {
			s.logger.Error(err.Error())
		}
	}()
	packetConnection = s.shadower.Shadow(packetConnection)

	NATMap := natmap{
		remoteAddressToConnection: make(map[string]net.PacketConn),
		timeNow:                   s.timeNow,
	}

	buffer := make([]byte, bufferSize)

	s.logger.Info("listening UDP on " + address)
	for {
		bytesRead, remoteAddress, err := packetConnection.ReadFrom(buffer)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}
			s.logger.Error("cannot read from UDP buffer: " + err.Error())
			continue
		}

		targetAddress, err := socks.ExtractAddress(buffer[:bytesRead])
		if err != nil {
			s.logger.Error("cannot extract target address: " + err.Error())
			continue
		}

		targetUDPAddress, err := net.ResolveUDPAddr("udp", targetAddress.String())
		if err != nil {
			s.logger.Error("cannot resolve target UDP address: " + err.Error())
			continue
		}

		payload := buffer[len(targetAddress):bytesRead]

		connection := NATMap.Get(remoteAddress.String())
		if connection == nil {
			s.logger.Info("UDP proxying " + remoteAddress.String() + " to " + targetAddress.String())
			connection, err = net.ListenPacket("udp", "")
			if err != nil {
				s.logger.Info("cannot listen to UDP packet: " + err.Error())
				continue
			}
			NATMap.Set(remoteAddress.String(), connection)
			go NATMap.Handle(remoteAddress, packetConnection, connection, s.logger)
		}

		if _, err := connection.WriteTo(payload, targetUDPAddress); err != nil {
			// accept only UDPAddr despite the signature
			s.logger.Error("cannot write to UDP address " + targetUDPAddress.String() + ": " + err.Error())
		}
	}
}
