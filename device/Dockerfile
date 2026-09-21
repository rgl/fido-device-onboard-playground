# syntax=docker.io/docker/dockerfile:1.27

# see https://github.com/fido-device-onboard/go-fdo-client/releases
FROM golang:1.27.1-trixie AS build-fdo-client
WORKDIR /src
RUN curl -L https://github.com/fido-device-onboard/go-fdo-client/archive/refs/tags/v1.0.0.tar.gz | tar xz --strip-components=1
RUN make build

FROM golang:1.27.1-trixie AS build
WORKDIR /src
COPY go.* ./
RUN go mod download
COPY *.go ./
RUN go build -o device -ldflags="-s -w" -trimpath

FROM debian:trixie-slim
RUN <<EOF
#!/bin/bash
set -euxo pipefail
apt-get update
apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    iproute2 \
    openssl \
    procps \
    sqlite3
rm -rf /var/lib/apt/lists/*
EOF
COPY --from=build-fdo-client /src/go-fdo-client /usr/local/bin/go-fdo-client
COPY --from=build /src/device /usr/local/bin/device
ENTRYPOINT ["device"]
