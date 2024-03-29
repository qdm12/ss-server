package udp

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/qdm12/ss-server/internal/core"
	"github.com/qdm12/ss-server/internal/filter"
	"github.com/qdm12/ss-server/internal/socks"
)

func NewServer(settings Settings, logger Logger) (s *Server, err error) {
	settings.SetDefaults()

	udpPacketCipher, err := core.NewUDPPacketCipher(
		settings.CipherName, *settings.Password, filter.NewBloomRing())
	if err != nil {
		return nil, err
	}
	return &Server{
		address:      *settings.Address,
		logAddresses: *settings.LogAddresses,
		logger:       logger,
		timeNow:      time.Now,
		shadower:     udpPacketCipher,
	}, nil
}

type Server struct {
	address      string
	logAddresses bool
	logger       Logger
	timeNow      func() time.Time
	shadower     *core.UDPPacketCipher
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

	s.logger.Info("listening UDP on " + packetConnection.LocalAddr().String())
	for {
		bytesRead, remoteAddress, err := packetConnection.ReadFrom(buffer)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}

			err = fmt.Errorf("reading packet: %w", err)
			if remoteAddress != nil {
				err = fmt.Errorf("connection from %s: %w", remoteAddress, err)
			}
			s.logger.Error(err.Error())
			continue
		}

		err = handleIncomingData(packetConnection, remoteAddress,
			buffer, bytesRead, &NATMap, s.logger, s.logAddresses)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}

			if remoteAddress != nil {
				err = fmt.Errorf("connection from %s: %w", remoteAddress, err)
			}

			s.logger.Error(err.Error())
		}
	}
}

func handleIncomingData(packetConnection net.PacketConn, remoteAddress net.Addr,
	buffer []byte, bytesRead int, natMap *natmap, logger Logger,
	logAddresses bool) (err error) {
	targetAddress, err := socks.ExtractAddress(buffer[:bytesRead])
	if err != nil {
		return fmt.Errorf("extracting SOCKS target address: %w", err)
	}

	targetUDPAddress, err := net.ResolveUDPAddr("udp", targetAddress.String())
	if err != nil {
		return fmt.Errorf("resolving target address: %w", err)
	}

	payload := buffer[len(targetAddress):bytesRead]

	connection := natMap.Get(remoteAddress.String())
	if connection == nil {
		if logAddresses {
			logger.Info("UDP proxying " + remoteAddress.String() + " to " + targetAddress.String())
		}

		connection, err = net.ListenPacket("udp", "")
		if err != nil {
			return fmt.Errorf("creating packet listener: %w", err)
		}
		natMap.Set(remoteAddress.String(), connection)
		go natMap.Handle(remoteAddress, packetConnection, connection)
	}

	_, err = connection.WriteTo(payload, targetUDPAddress)
	if err != nil {
		// accept only UDPAddr despite the signature
		return fmt.Errorf("writing payload to address %s: %w", targetUDPAddress, err)
	}

	return nil
}
