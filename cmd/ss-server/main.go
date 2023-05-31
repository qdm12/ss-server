package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/qdm12/log"
	"github.com/qdm12/ss-server/internal/config/settings"
	"github.com/qdm12/ss-server/internal/config/sources/env"
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

	reader := env.New()

	logger := log.New()

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

func _main(ctx context.Context, logger Logger, settingsSource ReaderInterface) error {
	settings, err := settingsSource.Read()
	if err != nil {
		return fmt.Errorf("reading settings: %w", err)
	}
	settings.SetDefaults()

	err = settings.Validate()
	if err != nil {
		return fmt.Errorf("validating settings: %w", err)
	}

	logger.Patch(log.SetLevel(*settings.LogLevel))

	serverSettings := tcpudp.Settings{
		Address:    ":" + fmt.Sprint(*settings.Port),
		CipherName: settings.CipherName,
		Password:   settings.Password,
	}

	server, err := tcpudp.NewServer(serverSettings, logger)
	if err != nil {
		return err
	}

	if *settings.Profiling {
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
	log.LoggerPatcher
}

type ReaderInterface interface {
	Read() (settings settings.Settings, err error)
}
