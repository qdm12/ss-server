package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/qdm12/ss-server/internal/log"
	"github.com/qdm12/ss-server/internal/profiling"
	"github.com/qdm12/ss-server/pkg"
)

//nolint: gochecknoglobals
var (
	BuildDate = "unknown date"
	VcsRef    = "unknown ref"
	Version   = "unknown"
)

func main() {
	ctx := context.Background()
	environ := os.Environ()
	os.Exit(_main(ctx, environ))
}

func _main(ctx context.Context, environment []string) int { //nolint:unparam
	cipherName := "chacha20-ietf-poly1305"
	password := "password"
	port := "8388"
	logLevel := "INFO"
	doProfiling := false
	for _, envVariable := range environment {
		slice := strings.Split(envVariable, "=")
		key, value := slice[0], slice[1]
		switch key {
		case "PASSWORD":
			password = value
		case "PORT":
			port = value
		case "CIPHER":
			cipherName = value
		case "LOG_LEVEL":
			logLevel = strings.ToUpper(value)
		case "PROFILING":
			if strings.ToLower(value) == "on" {
				doProfiling = true
			}
		}
	}

	logger := log.NewLogger("", log.Level(logLevel))

	logger.Info("Running version " + Version + " built on " + BuildDate + " (" + VcsRef + ")")

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
