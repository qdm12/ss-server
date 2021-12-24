# Shadowsocks server

Focuses on a Dockerized Shadowsocks server as well as giving an easy to use Go API to run a Shadowsocks server.

<img height="200" src="title.svg">

❓ Question, suggestion, request? ➡️ [Create an issue!](https://github.com/qdm12/ss-server/issues/new)

[![Build status](https://github.com/qdm12/ss-server/actions/workflows/ci.yml/badge.svg)](https://github.com/qdm12/ss-server/actions/workflows/ci.yml)

[![dockeri.co](https://dockeri.co/image/qmcgaw/ss-server)](https://hub.docker.com/r/qmcgaw/ss-server)

![Last release](https://img.shields.io/github/release/qdm12/ss-server?label=Last%20release)
![Last Docker tag](https://img.shields.io/docker/v/qmcgaw/ss-server?sort=semver&label=Last%20Docker%20tag)
[![Last release size](https://img.shields.io/docker/image-size/qmcgaw/ss-server?sort=semver&label=Last%20released%20image)](https://hub.docker.com/r/qmcgaw/ss-server/tags?page=1&ordering=last_updated)
![GitHub last release date](https://img.shields.io/github/release-date/qdm12/ss-server?label=Last%20release%20date)
![Commits since release](https://img.shields.io/github/commits-since/qdm12/ss-server/latest?sort=semver)

[![Latest size](https://img.shields.io/docker/image-size/qmcgaw/ss-server/latest?label=Latest%20image)](https://hub.docker.com/r/qmcgaw/ss-server/tags)

[![GitHub last commit](https://img.shields.io/github/last-commit/qdm12/ss-server.svg)](https://github.com/qdm12/ss-server/commits/main)
[![GitHub commit activity](https://img.shields.io/github/commit-activity/y/qdm12/ss-server.svg)](https://github.com/qdm12/ss-server/graphs/contributors)
[![GitHub closed PRs](https://img.shields.io/github/issues-pr-closed/qdm12/ss-server.svg)](https://github.com/qdm12/ss-server/pulls?q=is%3Apr+is%3Aclosed)
[![GitHub issues](https://img.shields.io/github/issues/qdm12/ss-server.svg)](https://github.com/qdm12/ss-server/issues)
[![GitHub closed issues](https://img.shields.io/github/issues-closed/qdm12/ss-server.svg)](https://github.com/qdm12/ss-server/issues?q=is%3Aissue+is%3Aclosed)

[![Lines of code](https://img.shields.io/tokei/lines/github/qdm12/ss-server)](https://github.com/qdm12/ss-server)
![Code size](https://img.shields.io/github/languages/code-size/qdm12/ss-server)
![GitHub repo size](https://img.shields.io/github/repo-size/qdm12/ss-server)
![Go version](https://img.shields.io/github/go-mod/go-version/qdm12/ss-server)

[![MIT](https://img.shields.io/github/license/qdm12/ss-server)](https://github.com/qdm12/ss-server/master/LICENSE)
![Visitors count](https://visitor-badge.laobi.icu/badge?page_id=ss-server.readme)

## Docker

The Docker image is:

- Based on Scratch (no OS) for a total size of **3.53MB**
- Compatible with all the CPU architectures supported by Docker: `linux/amd64`, `linux/386`, `linux/arm64`, `linux/arm/v6`, `linux/arm/v7`, `linux/s390x`, `linux/ppc64le` and `linux/riscv64`
- Shadowsocks is implemented in Go and compiled statically using Go 1.17

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
    password := "password"
    settings := tcpudp.Settings{
        Address:    ":8388",
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
```

The call to `server.Listen(ctx, ":8388")` is blocking but you can run in a goroutine and cancel the context `ctx` when you want to stop the server.

### TCP only and UDP only

API for the TCP only and UDP only are almost the same, with the difference that they return an error on exit.

- [TCP only example](examples/tcp/main.go)
- [UDP only example](examples/udp/main.go)

### Docker entrypoint

Have also a look at the [cmd/ss-server/main.go](cmd/ss-server/main.go) which is quite straight forward to understand but uses a bit more complex asynchronous parts.

### Testing

Interfaces are defined such as `github.com/qdm12/ss-server/pkg/tcpudp.Listener`:

```go
type Listener interface {
    Listen(ctx context.Context, address string) (err error)
}
```

You can generate mocks in your codebase with for example [github.com/golang/mock](github.com/golang/mock) from these interfaces.

[For example](examples/test/main_test.go):

```go
package main

import (
    "context"
    "testing"

    "github.com/golang/mock/gomock"
)

//go:generate mockgen -destination=mock_tcpudp_listener_test.go -package $GOPACKAGE github.com/qdm12/ss-server/pkg/tcpudp Listener

func Test_Mytest(t *testing.T) {
    t.Parallel()
    ctrl := gomock.NewController(t)
    defer ctrl.Finish() // for Go < 1.14

    ctx := context.Background()

    server := NewMockListener(ctrl)
    server.EXPECT().Listen(ctx, ":8388").Return(nil)

    err := server.Listen(ctx, ":8388")
    if err != nil {
        t.Error("not expecting an error")
    }
}
```

In this example, you can use `go generate ./...` to re-generate the mock from the `Listener` interface.

## On demand

- SIP003 plugins

## TODOS

- Support hex raw keys instead of passwords
- Entrypoint message
- Prometheus stats
- Docker healthcheck + healthcheck endpoint (i.e. for K8s)
