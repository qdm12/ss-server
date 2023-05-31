package tcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/qdm12/ss-server/internal/core"
	"github.com/qdm12/ss-server/internal/filter"
	"github.com/qdm12/ss-server/internal/socks"
)

func NewServer(settings Settings, logger Logger) (s *Server, err error) {
	settings.SetDefaults()

	tcpStreamCipher, err := core.NewTCPStreamCipher(
		settings.CipherName, *settings.Password, filter.NewBloomRing())
	if err != nil {
		return nil, err
	}
	return &Server{
		address:      settings.Address,
		logAddresses: *settings.LogAddresses,
		logger:       logger,
		timeNow:      time.Now,
		shadower:     tcpStreamCipher,
	}, nil
}

type Server struct {
	address      string
	logAddresses bool
	logger       Logger
	timeNow      func() time.Time
	shadower     *core.TCPStreamCipher
}

// Listen listens for incoming connections.
func (s *Server) Listen(ctx context.Context) (err error) {
	listenConfig := net.ListenConfig{}
	listener, err := listenConfig.Listen(ctx, "tcp", s.address)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		if err := listener.Close(); err != nil {
			s.logger.Error(err.Error())
		}
	}()
	s.logger.Info("listening TCP on " + s.address)
	for {
		connection, err := listener.Accept()
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
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

func (s *Server) handleConnection(connection net.Conn) {
	defer connection.Close()
	shadowedConnection := s.shadower.Shadow(connection)
	defer shadowedConnection.Close()

	targetAddress, err := socks.ReadAddress(shadowedConnection)
	if err != nil {
		s.logger.Error("cannot obtain target address: " + err.Error())
		if _, err := io.Copy(io.Discard, connection); err != nil {
			s.logger.Error(err.Error())
		}
		return
	}

	rightConnection, err := net.Dial("tcp", targetAddress.String())
	if err != nil {
		s.logger.Error("cannot connect to target address " + targetAddress.String() + ": " + err.Error())
		return
	}
	defer rightConnection.Close()

	if s.logAddresses {
		s.logger.Info("TCP proxying " + connection.RemoteAddr().String() + " to " + targetAddress.String())
	}

	if err := relay(shadowedConnection, rightConnection, s.timeNow); err != nil {
		var netErr net.Error
		if ok := errors.As(err, &netErr); ok && netErr.Timeout() {
			s.logger.Debug("TCP relay error: " + err.Error())
			return // ignore i/o timeout
		}
		s.logger.Error("TCP relay error: " + err.Error())
	}
}
