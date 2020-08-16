package tcp

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/qdm12/ss-server/internal/core"
	"github.com/qdm12/ss-server/internal/filter"
	"github.com/qdm12/ss-server/internal/log"
	"github.com/qdm12/ss-server/internal/socks"
)

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE.go . Server

type Server interface {
	Listen(ctx context.Context, address string) (err error)
}

func NewServer(cipherName, password string, logger log.Logger) (s Server, err error) {
	tcpStreamCipher, err := core.NewTCPStreamCipher(cipherName, password, filter.NewSaltFilter())
	if err != nil {
		return nil, err
	}
	return &server{
		logger:          logger,
		timeNow:         time.Now,
		tcpStreamCipher: tcpStreamCipher,
	}, nil
}

type server struct {
	logger          log.Logger
	timeNow         func() time.Time
	tcpStreamCipher core.TCPStreamCipher
}

// Listen listens on the address given for incoming connections
func (s *server) Listen(ctx context.Context, address string) (err error) {
	listenConfig := net.ListenConfig{}
	listener, err := listenConfig.Listen(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("cannot listen on %s: %w", address, err)
	}
	go func() {
		<-ctx.Done()
		if err := listener.Close(); err != nil {
			s.logger.Error(err.Error())
		}
	}()
	s.logger.Info(fmt.Sprintf("listening TCP on %s", address))
	for {
		connection, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			s.logger.Error("cannot accept connection on TCP listener: " + err.Error())
			continue
		}
		tcpConnection, ok := connection.(*net.TCPConn)
		if !ok {
			s.logger.Error(fmt.Sprintf("connection is not TCP: %s", connection))
			continue
		}
		if err := tcpConnection.SetKeepAlive(true); err != nil {
			s.logger.Error("cannot set keep-alive for TCP connection: " + err.Error())
			continue
		}
		go s.handleConnection(connection)
	}
}

func (s *server) handleConnection(connection net.Conn) {
	defer connection.Close()
	connection = s.tcpStreamCipher.Shadow(connection)

	targetAddress, err := socks.ReadAddress(connection)
	if err != nil {
		s.logger.Error("cannot obtain target address: " + err.Error())
		return
	}

	rightConnection, err := net.Dial("tcp", targetAddress.String())
	if err != nil {
		s.logger.Error(fmt.Sprintf("cannot connect to target address %q: %s", targetAddress, err))
		return
	}
	defer rightConnection.Close()
	rightTCPConnection, ok := rightConnection.(*net.TCPConn)
	if !ok {
		s.logger.Error(fmt.Sprintf("right connection is not TCP: %s", rightConnection))
		return
	}
	if err := rightTCPConnection.SetKeepAlive(true); err != nil {
		s.logger.Error("cannot set keep-alive for 'right' TCP connection: " + err.Error())
		return
	}

	s.logger.Info(fmt.Sprintf("TCP proxying %s to %s", connection.RemoteAddr(), targetAddress))
	if err := relay(connection, rightConnection, s.timeNow); err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			s.logger.Debug("TCP relay error: " + err.Error())
			return // ignore i/o timeout
		}
		s.logger.Error("TCP relay error: " + err.Error())
	}
}
