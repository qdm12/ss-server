# Shadowsocks server

Focuses on a Dockerized Shadowsocks server as well as giving an easy to use Go API to run a Shadowsocks server.

## Docker

The Docker image is:

- Based on Scratch (no OS) for a total size of **3.53MB**
- Compatible with `amd64`, `386` and all `arm` (i.e. raspberry pis)
- Shadowsocks is implemented in Go and compiled statically using Go 1.15

Run the container interactively to try it out

```sh
docker run -it --rm -p 8388:8388/tcp -p 8388:8388/udp -e PASSWORD=password qmcgaw/ss-server
```

Or use docker-compose.yml

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

	"github.com/qdm12/ss-server/pkg"
)

func main() {
	logger := &logger{}
	server, err := pkg.NewServer("aes-256-gcm", "password", logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	ctx := context.Background()
	server.Listen(ctx, "0.0.0.0:8388") // blocking call, can be run in a goroutine
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
	"github.com/qdm12/ss-server/pkg/mock_pkg"
)

func Test(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // for Go < 1.14
	server := mock_pkg.NewMockServer(ctrl)
	server.EXPECT().Listen(context.Background(), "0.0.0.0:8388")
	// more of your test using server
}
```

## TODOS

- Prometheus stats
- Plugins (on demand)
- Docker healthcheck + healthcheck endpoint (i.e. for K8s)
