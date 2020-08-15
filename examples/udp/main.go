package main

import (
	"context"
	"fmt"
	"os"

	"github.com/qdm12/ss-server/pkg/tcp"
)

func main() {
	logger := &logger{}
	server, err := tcp.NewServer("aes-256-gcm", "password", logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	ctx := context.Background()
	if err := server.Listen(ctx, "0.0.0.0:8388"); err != nil {
		logger.Error(err.Error())
	}
}

type logger struct{}

func (l *logger) Debug(s string) { fmt.Println("debug:", s) }
func (l *logger) Info(s string)  { fmt.Println("info:", s) }
func (l *logger) Error(s string) { fmt.Println("error:", s) }
