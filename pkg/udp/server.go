package udp

import (
	"context"
	"fmt"
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
	Listen(ctx context.Context) (err error)
}

func NewServer(settings Settings, logger log.Logger) (s *Server, err error) {
	settings.setDefaults()

	udpPacketCipher, err := core.NewUDPPacketCipher(
		settings.CipherName, settings.Password, filter.NewBloomRing())
	if err != nil {
		return nil, err
	}
	return &Server{
		address:      settings.Address,
		logAddresses: settings.LogAddresses,
		logger:       logger,
		timeNow:      time.Now,
		shadower:     udpPacketCipher,
	}, nil
}

type Server struct {
	address      string
	logAddresses bool
	logger       log.Logger
	timeNow      func() time.Time
	shadower     core.PacketConnShadower
}

// Listen listens for encrypted packets and does UDP NATing.
func (s *Server) Listen(ctx context.Context) (err error) {
	listenConfig := net.ListenConfig{}
	packetConnection, err := listenConfig.ListenPacket(ctx, "udp", s.address)
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

	s.logger.Info("listening UDP on " + s.address)
	for {
		err = handleIncomingData(packetConnection, buffer,
			&NATMap, s.logger, s.logAddresses)
		if err != nil {
			ctxErr := ctx.Err()
			if ctxErr != nil {
				return ctxErr
			}
			s.logger.Error(err.Error())
		}
	}
}

func handleIncomingData(packetConnection net.PacketConn, buffer []byte,
	natMap *natmap, logger log.Logger, logAddresses bool) (err error) {
	bytesRead, remoteAddress, err := packetConnection.ReadFrom(buffer)
	if err != nil {
		return fmt.Errorf("cannot read from UDP buffer: %w", err)
	}

	targetAddress, err := socks.ExtractAddress(buffer[:bytesRead])
	if err != nil {
		return fmt.Errorf("cannot extract target address: %w", err)
	}

	targetUDPAddress, err := net.ResolveUDPAddr("udp", targetAddress.String())
	if err != nil {
		return fmt.Errorf("cannot resolve target UDP address: %w", err)
	}

	payload := buffer[len(targetAddress):bytesRead]

	connection := natMap.Get(remoteAddress.String())
	if connection == nil {
		if logAddresses {
			logger.Info("UDP proxying " + remoteAddress.String() + " to " + targetAddress.String())
		}

		connection, err = net.ListenPacket("udp", "")
		if err != nil {
			return fmt.Errorf("cannot listen to UDP packet: %w", err)
		}
		natMap.Set(remoteAddress.String(), connection)
		go natMap.Handle(remoteAddress, packetConnection, connection, logger)
	}

	_, err = connection.WriteTo(payload, targetUDPAddress)
	if err != nil {
		// accept only UDPAddr despite the signature
		return fmt.Errorf("cannot write to UDP address %s: %w", targetUDPAddress, err)
	}

	return nil
}
