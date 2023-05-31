package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/qdm12/ss-server/internal/env"
	"github.com/qdm12/ss-server/internal/log"
	"github.com/qdm12/ss-server/internal/profiling"
	"github.com/qdm12/ss-server/pkg/tcpudp"
)

//nolint:gochecknoglobals
var (
	version = "unknown"
	date    = "unknown"
	commit  = "unknown"
)

func main() {
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	environ := os.Environ()
	reader := env.NewReader(environ)
	logLevel := reader.LogLevel()

	logger := log.New(logLevel, os.Stdout)

	logger.Info("Running version " + version + " built on " + date + " (" + commit + ")")

	errorCh := make(chan error)
	go func() {
		errorCh <- _main(ctx, logger, reader)
	}()

	select {
	case <-ctx.Done():
		logger.Warn("Caught OS signal, shutting down\n")
		stop()
	case err := <-errorCh:
		close(errorCh)
		if err == nil { // expected exit such as healthcheck
			os.Exit(0)
		}
		logger.Error(err.Error())
	}

	const shutdownGracePeriod = 5 * time.Second
	timer := time.NewTimer(shutdownGracePeriod)
	select {
	case <-errorCh:
		if !timer.Stop() {
			<-timer.C
		}
		logger.Info("Shutdown successful")
	case <-timer.C:
		logger.Warn("Shutdown timed out")
	}

	os.Exit(1)
}

func _main(ctx context.Context, logger Logger, reader ReaderInterface) error {
	cipherName, password, port, doProfiling :=
		reader.CipherName(), reader.Password(), reader.Port(), reader.Profiling()

	settings := tcpudp.Settings{
		Address:    ":" + port,
		CipherName: cipherName,
		Password:   &password,
	}

	server, err := tcpudp.NewServer(settings, logger)
	if err != nil {
		return err
	}

	if doProfiling {
		logger.Info("profiling server listening on :6060")
		onShutdownError := func(err error) { logger.Error(err.Error()) }
		profileServer := profiling.NewServer(onShutdownError)
		go func() {
			if err := profileServer.Run(ctx); err != nil {
				logger.Error(err.Error())
			}
		}()
	}

	return server.Listen(ctx)
}

type Logger interface {
	Debug(s string)
	Info(s string)
	Error(s string)
}

type ReaderInterface interface {
	CipherName() (cipherName string)
	Password() (password string)
	Port() (port string)
	LogLevel() (logLevel log.Level)
	Profiling() (profiling bool)
}
