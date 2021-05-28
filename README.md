# Shadowsocks server

Focuses on a Dockerized Shadowsocks server as well as giving an easy to use Go API to run a Shadowsocks server.

<img height="200" src="title.svg?sanitize=true">

❓ Question, suggestion, request? ➡️ [Create an issue!](https://github.com/qdm12/ss-server/issues/new)

[![Build status](https://github.com/qdm12/ss-server/workflows/Buildx%20latest/badge.svg)](https://github.com/qdm12/ss-server/actions?query=workflow%3A%22Buildx+latest%22)
[![Docker Pulls](https://img.shields.io/docker/pulls/qmcgaw/ss-server.svg)](https://hub.docker.com/r/qmcgaw/ss-server)
[![Docker Stars](https://img.shields.io/docker/stars/qmcgaw/ss-server.svg)](https://hub.docker.com/r/qmcgaw/ss-server)

[![GitHub last commit](https://img.shields.io/github/last-commit/qdm12/ss-server.svg)](https://github.com/qdm12/ss-server/issues)
[![GitHub commit activity](https://img.shields.io/github/commit-activity/y/qdm12/ss-server.svg)](https://github.com/qdm12/ss-server/issues)
[![GitHub issues](https://img.shields.io/github/issues/qdm12/ss-server.svg)](https://github.com/qdm12/ss-server/issues)

[![Image size](https://images.microbadger.com/badges/image/qmcgaw/ss-server.svg)](https://microbadger.com/images/qmcgaw/ss-server)
[![Image version](https://images.microbadger.com/badges/version/qmcgaw/ss-server.svg)](https://microbadger.com/images/qmcgaw/ss-server)
[![Join Slack channel](https://img.shields.io/badge/slack-@qdm12-yellow.svg?logo=slack)](https://join.slack.com/t/qdm12/shared_invite/enQtOTE0NjcxNTM1ODc5LTYyZmVlOTM3MGI4ZWU0YmJkMjUxNmQ4ODQ2OTAwYzMxMTlhY2Q1MWQyOWUyNjc2ODliNjFjMDUxNWNmNzk5MDk)

## Docker

The Docker image is:

- Based on Scratch (no OS) for a total size of **3.53MB**
- Compatible with `amd64`, `386` and all `arm` (i.e. raspberry pis)
- Shadowsocks is implemented in Go and compiled statically using Go 1.16

Run the container interactively to try it out

```sh
docker run -it --rm -p 8388:8388/tcp -p 8388:8388/udp -e PASSWORD=password qmcgaw/ss-server
```

Or use docker-compose.yml with `docker-compose up -d`

```yml
version: "3.7"
services:
  shadowsocks:
      image: qmcgaw/ss-server
      container_name: shadowsocks
      network_mode: bridge
      ports:
          - 8388:8388/tcp
          - 8388:8388/udp
      environment:
          - PASSWORD=password
          - TZ=
      restart: always
```

The environment variables are:

| Name | Default | Possible values | Description |
| --- | --- | --- | --- |
| `PASSWORD` |  | Any password | Your password |
| `PORT` | `8388` | `1024` to `65535` | Internal listening port |
| `LOG_LEVEL` | `INFO` | `INFO`, `ERROR`, `DEBUG` | Log level |
| `CIPHER` | `chacha20-ietf-poly1305` | `chacha20-ietf-poly1305`, `aes-128-gcm`, `aes-256-gcm` | Cipher to use |
| `TZ` |  | Timezone, i.e. `America/Montreal` | Timezone for log times display |
| `PROFILING` | `off` | `on` or `off` | Enable the Go pprof http server on `:6060` |

## Go API

This repository was designed such that it is easy to integrate and launch safely a Shadowsocks server from an existing Go program.

### TCP+UDP example

[Source file](examples/tcp-udp/main.go)

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/qdm12/ss-server/pkg/tcpudp"
)

func main() {
	logger := &logger{}
	server, err := tcpudp.NewServer("aes-256-gcm", "password", logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	ctx := context.Background()
	err = server.Listen(ctx, "0.0.0.0:8388") // blocking call, can be run in a goroutine
	if err != nil {
		logger.Error(err.Error())
	}
}

type logger struct{}

func (l *logger) Debug(s string) { fmt.Println("debug:", s) }
func (l *logger) Info(s string)  { fmt.Println("info:", s) }
func (l *logger) Error(s string) { fmt.Println("error:", s) }
```

The call to `server.Listen(ctx, "0.0.0.0:8388")` is blocking but you can run in a goroutine and cancel the context `ctx` when you want to stop the server.

### TCP only and UDP only

API for the TCP only and UDP only are almost the same, with the difference that they return an error on exit.

- [TCP only example](examples/tcp/main.go)
- [UDP only example](examples/udp/main.go)

### Docker entrypoint

Have also a look at the [cmd/ss-server/main.go](cmd/ss-server/main.go) which is quite straight forward to understand but uses a bit more complex asynchronous parts.

### Testing

Mocks are generated and committed to source control for each `Server` interface so you can directly use them with `gomock` for you tests. [For example](examples/test/main_test.go):

```go
package main

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/qdm12/ss-server/pkg/tcpudp/mock_tcpudp"
)

func Test(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // for Go < 1.14
	server := mock_tcpudp.NewMockServer(ctrl)
	server.EXPECT().Listen(context.Background(), "0.0.0.0:8388")
	// more of your test using server
}
```

## On demand

- More CPU architectures, for example `s390x`
- SIP003 plugins

## TODOS

- Support hex raw keys instead of passwords
- Readme svg icon
- Entrypoint message
- Prometheus stats
- Docker healthcheck + healthcheck endpoint (i.e. for K8s)
