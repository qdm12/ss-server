package main

import (
	"context"
	"fmt"
	"os"

	"github.com/qdm12/ss-server/pkg/udp"
)

func main() {
	logger := &logger{}
	settings := udp.Settings{
		Address:    ":8388",
		CipherName: "aes-256-gcm",
		Password:   "password",
	}
	server, err := udp.NewServer(settings, logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	ctx := context.Background()
	if err := server.Listen(ctx); err != nil {
		logger.Error(err.Error())
	}
}

type logger struct{}

func (l *logger) Debug(s string) { fmt.Println("debug:", s) }
func (l *logger) Info(s string)  { fmt.Println("info:", s) }
func (l *logger) Error(s string) { fmt.Println("error:", s) }
