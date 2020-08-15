ARG ALPINE_VERSION=3.12
ARG GO_VERSION=1.15

FROM alpine:${ALPINE_VERSION} AS alpine
RUN apk --update add tzdata

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder
ARG GOLANGCI_LINT_VERSION=v1.30.0
RUN apk --update add git
ENV CGO_ENABLED=0
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s ${GOLANGCI_LINT_VERSION}
WORKDIR /tmp/gobuild
COPY .golangci.yml .
COPY go.mod go.sum ./
RUN go mod download 2>&1
COPY . .
RUN go test ./...
RUN golangci-lint run --timeout=10m
ARG BUILD_DATE=unknown
ARG VCS_REF=local
ARG VERSION=local
RUN go build -o app -trimpath -ldflags="-s -w \
    -X 'main.BuildDate=$BUILD_DATE' \
    -X 'main.VcsRef=$VCS_REF' \
    -X 'main.Version=$VERSION'" \
    cmd/ss-server/main.go

FROM scratch
ARG VERSION
ARG BUILD_DATE
ARG VCS_REF
LABEL \
    org.opencontainers.image.authors="quentin.mcgaw@gmail.com" \
    org.opencontainers.image.version=$VERSION \
    org.opencontainers.image.created=$BUILD_DATE \
    org.opencontainers.image.revision=$VCS_REF \
    org.opencontainers.image.url="https://github.com/qdm12/ss-server" \
    org.opencontainers.image.documentation="https://github.com/qdm12/ss-server/internal/blob/master/README.md" \
    org.opencontainers.image.source="https://github.com/qdm12/ss-server" \
    org.opencontainers.image.title="ss-server" \
    org.opencontainers.image.description="Shadowsocks server written in Go, aimed for Docker containers"
COPY --from=alpine --chown=1000 /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ= \
    LOG_LEVEL=INFO \
    PORT=8388 \
    CIPHER=chacha20-ietf-poly1305
ENTRYPOINT ["/ss-server"]
# HEALTHCHECK --interval=10s --timeout=5s --start-period=5s --retries=2 CMD ["/app","healthcheck"]
USER 1000
COPY --from=builder --chown=1000 /tmp/gobuild/app /ss-server
