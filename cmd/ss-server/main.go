package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/qdm12/ss-server/internal/log"
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
		}
	}

	logger := log.NewLogger("", log.Level(logLevel))

	logger.Info(fmt.Sprintf("Running version %s built on %s (%s)", Version, BuildDate, VcsRef))

	server, err := pkg.NewServer(cipherName, password, logger)
	if err != nil {
		logger.Error(err.Error())
		return 1
	}

	ctx, cancel := context.WithCancel(ctx)

	serverExited := make(chan struct{})
	go func() {
		server.Listen(ctx, "0.0.0.0:"+port)
		close(serverExited)
	}()

	OSSignals := make(chan os.Signal, 1)
	signal.Notify(OSSignals, syscall.SIGINT, syscall.SIGTERM)
	signal := <-OSSignals
	logger.Info("Received OS signal " + signal.String())
	cancel()
	<-serverExited
	return 1
}
