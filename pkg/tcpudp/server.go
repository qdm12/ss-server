package tcpudp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qdm12/ss-server/internal/log"
	"github.com/qdm12/ss-server/pkg/tcp"
	"github.com/qdm12/ss-server/pkg/udp"
)

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE . Server

type Server interface {
	Listen(ctx context.Context, address string) (err error)
}

type server struct {
	timeNow   func() time.Time
	tcpServer tcp.Server
	udpServer udp.Server
	logger    log.Logger
}

func NewServer(cipherName, password string, logger log.Logger) (s Server, err error) {
	tcpServer, err := tcp.NewServer(cipherName, password, logger)
	if err != nil {
		return nil, err
	}
	udpServer, err := udp.NewServer(cipherName, password, logger)
	if err != nil {
		return nil, err
	}
	return &server{
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

func (s *server) Listen(ctx context.Context, address string) (err error) {
	ctx, cancel := context.WithCancel(ctx)

	serversRunning := map[string]struct{}{
		"TCP server": {},
		"UDP server": {},
	}
	exited := make(chan string)

	// Launch TCP and UDP servers
	errorCh := make(chan error)
	go func() {
		udpErr := s.udpServer.Listen(ctx, address)
		if ctx.Err() == nil && udpErr != nil {
			errorCh <- fmt.Errorf("%w: %s", ErrUDPServer, udpErr)
		}
		exited <- "UDP server"
	}()
	go func() {
		tcpErr := s.tcpServer.Listen(ctx, address)
		if ctx.Err() == nil && tcpErr != nil {
			errorCh <- fmt.Errorf("%w: %s", ErrTCPServer, tcpErr)
		}
		exited <- "TCP server"
	}()

	select {
	case err = <-errorCh: // unexpected error
	case <-ctx.Done(): // parent canceled on us
	}
	cancel() // stop the other server if an error occurred
	close(errorCh)

	const shutdownGracePeriod = 500 * time.Millisecond
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), shutdownGracePeriod)
	defer timeoutCancel()

	for len(serversRunning) > 0 {
		select {
		case serverExited := <-exited:
			s.logger.Info(serverExited + " exited")
			delete(serversRunning, serverExited)
		case <-timeoutCtx.Done():
			for serverNotExited := range serversRunning {
				s.logger.Error(serverNotExited + " did not exit during the " +
					shutdownGracePeriod.String() + " shutdown grace period")
			}
			return err
		}
	}
	return err
}
