package pkg

import (
	"context"
	"fmt"
	"time"

	"github.com/qdm12/ss-server/internal/log"
	"github.com/qdm12/ss-server/pkg/tcp"
	"github.com/qdm12/ss-server/pkg/udp"
)

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE.go . Server

type Server interface {
	Listen(ctx context.Context, address string)
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

func (s *server) Listen(ctx context.Context, address string) {
	ctx, cancel := context.WithCancel(ctx)

	serversRunning := map[string]struct{}{
		"TCP server": {},
		"UDP server": {},
	}
	exited := make(chan string)

	// Launch TCP and UDP servers
	go func() {
		if err := s.udpServer.Listen(ctx, address); err != nil {
			s.logger.Error(fmt.Sprintf("UDP server exited: %s", err))
		}
		cancel()
		exited <- "UDP server"
	}()
	go func() {
		if err := s.tcpServer.Listen(ctx, address); err != nil {
			s.logger.Error(fmt.Sprintf("TCP server exited: %s", err))
		}
		cancel()
		exited <- "TCP server"
	}()

	<-ctx.Done()

	const shutdownGracePeriod = 500 * time.Millisecond
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), shutdownGracePeriod)
	defer timeoutCancel()

	for len(serversRunning) > 0 {
		select {
		case serverExited := <-exited:
			s.logger.Info(fmt.Sprintf("%s exited successfully", serverExited))
			delete(serversRunning, serverExited)
		case <-timeoutCtx.Done():
			for serverNotExited := range serversRunning {
				s.logger.Error(fmt.Sprintf("%s did not exit during the %s shutdown grace period", serverNotExited, shutdownGracePeriod))
			}
			return
		}
	}
}
