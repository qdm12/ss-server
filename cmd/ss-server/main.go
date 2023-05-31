package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/qdm12/gosplash"
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
	buildInfo := BuildInformation{
		Version: version,
		Commit:  commit,
		Date:    date,
	}

	background := context.Background()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(background)

	logger := log.New()
	envReader := env.New()

	errorCh := make(chan error)
	go func() {
		errorCh <- _main(ctx, buildInfo, logger, envReader)
	}()

	var err error
	select {
	case signal := <-signalCh:
		fmt.Println("")
		logger.Warn("Caught OS signal " + signal.String() + ", shutting down")
		cancel()
	case err = <-errorCh:
		close(errorCh)
		if err == nil { // expected exit such as healthcheck
			os.Exit(0)
		}
		logger.Error(err.Error())
		cancel()
	}

	const shutdownGracePeriod = 5 * time.Second
	timer := time.NewTimer(shutdownGracePeriod)
	select {
	case shutdownErr := <-errorCh:
		if !timer.Stop() {
			<-timer.C
		}
		if shutdownErr != nil {
			logger.Warnf("Shutdown not completed gracefully: %s", shutdownErr)
			os.Exit(1)
		}

		logger.Info("Shutdown successful")
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	case <-timer.C:
		logger.Warn("Shutdown timed out")
		os.Exit(1)
	case signal := <-signalCh:
		logger.Warn("Caught OS signal " + signal.String() + ", forcing shut down")
		os.Exit(1)
	}
}

func _main(ctx context.Context, buildInfo BuildInformation,
	logger Logger, settingsSource ReaderInterface) error {
	splashSettings := gosplash.Settings{
		User:       "qdm12",
		Repository: "ss-server",
		Emails:     []string{"quentin.mcgaw@gmail.com"},
		Version:    buildInfo.Version,
		Commit:     buildInfo.Commit,
		BuildDate:  buildInfo.Date,
		// Sponsor information
		PaypalUser:    "qmcgaw",
		GithubSponsor: "qdm12",
	}
	for _, line := range gosplash.MakeLines(splashSettings) {
		fmt.Println(line)
	}

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
		Address:    *settings.Address,
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

type BuildInformation struct {
	Version string
	Commit  string
	Date    string
}
