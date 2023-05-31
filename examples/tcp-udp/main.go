package main

import (
	"context"
	"fmt"
	"os"

	"github.com/qdm12/ss-server/pkg/tcpudp"
)

func main() {
	logger := &logger{}
	password := "password"
	address := ":8388"
	settings := tcpudp.Settings{
		Address:    &address,
		CipherName: "aes-256-gcm",
		Password:   &password,
	}
	server, err := tcpudp.NewServer(settings, logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	ctx := context.Background()
	err = server.Listen(ctx) // blocking call, can be run in a goroutine
	if err != nil {
		logger.Error(err.Error())
	}
}

type logger struct{}

func (l *logger) Debug(s string) { fmt.Println("debug:", s) }
func (l *logger) Info(s string)  { fmt.Println("info:", s) }
func (l *logger) Error(s string) { fmt.Println("error:", s) }
