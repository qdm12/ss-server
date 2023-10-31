package tcpudp

import (
	"context"
	"errors"
	"time"

	"github.com/qdm12/ss-server/pkg/tcp"
	"github.com/qdm12/ss-server/pkg/udp"
)

type Server struct {
	timeNow   func() time.Time
	tcpServer *tcp.Server
	udpServer *udp.Server
	logger    Logger
}

func NewServer(settings Settings, logger Logger) (s *Server, err error) {
	settings.SetDefaults()

	tcpServer, err := tcp.NewServer(settings.TCP, logger)
	if err != nil {
		return nil, err
	}
	udpServer, err := udp.NewServer(settings.UDP, logger)
	if err != nil {
		return nil, err
	}
	return &Server{
		timeNow:   time.Now,
		tcpServer: tcpServer,
		udpServer: udpServer,
		logger:    logger,
	}, nil
}

var (
	ErrUDPServer = errors.New("UDP server crashed")
	ErrTCPServer = errors.New("TCP server crashed")
)

func (s *Server) Listen(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)

	tcpErrorCh := make(chan error)
	udpErrorCh := make(chan error)

	// Launch TCP and UDP servers
	go func() {
		udpErrorCh <- s.udpServer.Listen(ctx)
	}()
	go func() {
		tcpErrorCh <- s.tcpServer.Listen(ctx)
	}()

	select {
	case err = <-tcpErrorCh:
		s.logger.Info("TCP server exited: " + err.Error())
		cancel()
		close(tcpErrorCh)
		<-udpErrorCh
		s.logger.Info("UDP server exited")
		close(udpErrorCh)
	case err = <-udpErrorCh:
		s.logger.Info("UDP server exited: " + err.Error())
		cancel()
		close(udpErrorCh)
		<-tcpErrorCh
		s.logger.Info("TCP server exited")
		close(tcpErrorCh)
	case <-ctx.Done(): // parent canceled on us
		cancel() // for the linter
		<-udpErrorCh
		<-tcpErrorCh
		close(udpErrorCh)
		close(tcpErrorCh)
	}

	return err
}
