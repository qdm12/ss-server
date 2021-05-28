package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/qdm12/ss-server/internal/env"
	"github.com/qdm12/ss-server/internal/log"
	"github.com/qdm12/ss-server/internal/profiling"
	"github.com/qdm12/ss-server/pkg"
)

//nolint: gochecknoglobals
var (
	version = "unknown"
	date    = "unknown"
	commit  = "unknown"
)

func main() {
	ctx := context.Background()
	environ := os.Environ()
	os.Exit(_main(ctx, environ))
}

func _main(ctx context.Context, environ []string) int { //nolint:unparam
	reader := env.NewReader(environ)
	cipherName, password, port, logLevel, doProfiling :=
		reader.CipherName(), reader.Password(), reader.Port(),
		reader.LogLevel(), reader.Profiling()

	logger := log.NewLogger("", log.Level(logLevel))

	logger.Info("Running version " + version + " built on " + date + " (" + commit + ")")

	server, err := pkg.NewServer(cipherName, password, logger)
	if err != nil {
		logger.Error(err.Error())
		return 1
	}

	ctx, cancel := context.WithCancel(ctx)

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

	errorCh := make(chan error)
	go func() {
		errorCh <- server.Listen(ctx, ":"+port)
	}()

	OSSignals := make(chan os.Signal, 1)
	signal.Notify(OSSignals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errorCh:
		logger.Error(err.Error())
		cancel()
	case signal := <-OSSignals:
		logger.Info("Received OS signal " + signal.String())
		cancel()
		<-errorCh // wait for exit
	}
	close(errorCh)
	return 1
}
