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
		address:      *settings.Address,
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
	s.logger.Info("listening TCP on " + listener.Addr().String())
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
			s.logger.Error(fmt.Sprintf("connection from %s is not TCP: %s",
				connection.RemoteAddr(), connection))
			continue
		}
		if err := tcpConnection.SetKeepAlive(true); err != nil {
			s.logger.Error(fmt.Sprintf("cannot set keep-alive for TCP connection from %s: %s",
				connection.RemoteAddr(), err))
			continue
		}
		go s.handleConnectionAsync(connection)
	}
}

func (s *Server) handleConnectionAsync(connection net.Conn) {
	errs := s.handleConnection(connection)
	for _, err := range errs {
		s.logger.Error(fmt.Sprintf("connection from %s: %s", connection.RemoteAddr(), err))
	}
}

func (s *Server) handleConnection(connection net.Conn) (errs []error) {
	defer closeConnection("TCP connection", connection, &errs)

	shadowedConnection := s.shadower.Shadow(connection)
	defer closeConnection("shadowed TCP connection", shadowedConnection, &errs)

	targetAddress, err := socks.ReadAddress(shadowedConnection)
	if err != nil {
		errs = append(errs, fmt.Errorf("reading target address: %w", err))
		if _, err := io.Copy(io.Discard, connection); err != nil {
			errs = append(errs, fmt.Errorf("discarding connection data: %w", err))
		}
		return errs
	}

	rightConnection, err := net.Dial("tcp", targetAddress.String())
	if err != nil {
		errs = append(errs, fmt.Errorf("connecting to target address %s: %w", targetAddress, err))
		return errs
	}
	defer closeConnection("TCP connection to target address", rightConnection, &errs)

	if s.logAddresses {
		s.logger.Info("TCP proxying " + connection.RemoteAddr().String() + " to " + targetAddress.String())
	}

	if err := relay(shadowedConnection, rightConnection, s.timeNow); err != nil {
		var netErr net.Error
		if ok := errors.As(err, &netErr); ok && netErr.Timeout() {
			s.logger.Debug("TCP relay error: " + err.Error())
			return // ignore i/o timeout
		}
		errs = append(errs, fmt.Errorf("TCP relay error: %w", err))
	}

	return errs
}

func closeConnection(name string, conn io.Closer, errs *[]error) {
	err := conn.Close()
	if err != nil {
		err = fmt.Errorf("closing %s: %w", name, err)
		*errs = append(*errs, err)
	}
}
