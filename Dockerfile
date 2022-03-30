# Sets linux/amd64 in case it's not injected by older Docker versions
ARG BUILDPLATFORM=linux/amd64

ARG ALPINE_VERSION=3.15
ARG GO_VERSION=1.17
ARG XCPUTRANSLATE_VERSION=v0.6.0
ARG GOLANGCI_LINT_VERSION=v1.45.2

FROM --platform=${BUILDPLATFORM} qmcgaw/xcputranslate:${XCPUTRANSLATE_VERSION} AS xcputranslate
FROM --platform=${BUILDPLATFORM} qmcgaw/binpot:golangci-lint-${GOLANGCI_LINT_VERSION} AS golangci-lint

FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS base
ENV CGO_ENABLED=0
WORKDIR /tmp/gobuild
RUN apk --update add git g++
COPY --from=xcputranslate /xcputranslate /usr/local/bin/xcputranslate
COPY --from=golangci-lint /bin /go/bin/golangci-lint
COPY go.mod go.sum ./
RUN go mod download
COPY pkg/ ./pkg/
COPY cmd/ ./cmd/
COPY internal/ ./internal/

FROM base AS test
# Note on the go race detector:
# - we set CGO_ENABLED=1 to have it enabled
# - we installed g++ in the base stage to support the race detector
ENV CGO_ENABLED=1
ENTRYPOINT go test -race -coverpkg=./... -coverprofile=coverage.txt -covermode=atomic ./...

FROM base AS lint
COPY .golangci.yml ./
RUN golangci-lint run --timeout=10m

FROM base AS build
ARG TARGETPLATFORM
ARG VERSION=unknown
ARG CREATED="an unknown date"
ARG COMMIT=unknown
RUN GOARCH="$(xcputranslate translate -targetplatform=${TARGETPLATFORM} -field arch)" \
    GOARM="$(xcputranslate translate -targetplatform=${TARGETPLATFORM} -field arm)" \
    go build -trimpath -ldflags="-s -w \
    -X 'main.version=$VERSION' \
    -X 'main.date=$CREATED' \
    -X 'main.commit=$COMMIT' \
    " -o app cmd/ss-server/main.go

FROM scratch
ARG VERSION=unknown
ARG CREATED=unknown
ARG COMMIT=unknown
LABEL \
    org.opencontainers.image.authors="quentin.mcgaw@gmail.com" \
    org.opencontainers.image.version=$VERSION \
    org.opencontainers.image.created=$CREATED \
    org.opencontainers.image.revision=$COMMIT \
    org.opencontainers.image.url="https://github.com/qdm12/ss-server" \
    org.opencontainers.image.documentation="https://github.com/qdm12/ss-server/internal/blob/master/README.md" \
    org.opencontainers.image.source="https://github.com/qdm12/ss-server" \
    org.opencontainers.image.title="ss-server" \
    org.opencontainers.image.description="Shadowsocks server written in Go, aimed for Docker containers"
ENV TZ= \
    LOG_LEVEL=INFO \
    PORT=8388 \
    CIPHER=chacha20-ietf-poly1305
ENTRYPOINT ["/ss-server"]
# HEALTHCHECK --interval=10s --timeout=5s --start-period=5s --retries=2 CMD ["/app","healthcheck"]
USER 1000
COPY --from=build --chown=1000 /tmp/gobuild/app /ss-server
