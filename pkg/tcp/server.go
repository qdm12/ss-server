package tcp

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
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
	tcpStreamCipher, err := core.NewTCPStreamCipher(cipherName, password, filter.NewBloomRing())
	if err != nil {
		return nil, err
	}
	return &Server{
		logger:   logger,
		timeNow:  time.Now,
		shadower: tcpStreamCipher,
	}, nil
}

type Server struct {
	logger   log.Logger
	timeNow  func() time.Time
	shadower core.ConnShadower
}

// Listen listens on the address given for incoming connections.
func (s *Server) Listen(ctx context.Context, address string) (err error) {
	listenConfig := net.ListenConfig{}
	listener, err := listenConfig.Listen(ctx, "tcp", address)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		if err := listener.Close(); err != nil {
			s.logger.Error(err.Error())
		}
	}()
	s.logger.Info("listening TCP on " + address)
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
		if _, err := io.Copy(ioutil.Discard, connection); err != nil {
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

	s.logger.Info("TCP proxying " + connection.RemoteAddr().String() + " to " + targetAddress.String())
	if err := relay(shadowedConnection, rightConnection, s.timeNow); err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			s.logger.Debug("TCP relay error: " + err.Error())
			return // ignore i/o timeout
		}
		s.logger.Error("TCP relay error: " + err.Error())
	}
}
