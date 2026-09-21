# syntax=docker.io/docker/dockerfile:1.27

# see https://github.com/fido-device-onboard/go-fdo-server/releases
FROM golang:1.27.1-trixie AS build
WORKDIR /src
RUN curl -L https://github.com/fido-device-onboard/go-fdo-server/archive/refs/tags/v1.0.1.tar.gz | tar xz --strip-components=1
RUN make build

# see https://github.com/mmartinv/go-fdo-tool
# see https://github.com/mmartinv/go-fdo-tool/tree/7b2014534039611ed854178b7368dd2fcd210a2c # 2026-07-20T05:39:53.000Z
FROM golang:1.27.1-trixie AS build-go-fdo-tool
WORKDIR /src
RUN curl -L https://github.com/mmartinv/go-fdo-tool/archive/7b2014534039611ed854178b7368dd2fcd210a2c.tar.gz | tar xz --strip-components=1
RUN go build -o go-fdo-tool -ldflags="-w -s" -trimpath

FROM debian:trixie-slim
RUN <<EOF
#!/bin/bash
set -euxo pipefail
apt-get update
apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    iproute2 \
    procps \
    sqlite3
rm -rf /var/lib/apt/lists/*
EOF
COPY --from=build /src/go-fdo-server /usr/local/bin/go-fdo-server
COPY --from=build-go-fdo-tool /src/go-fdo-tool /usr/local/bin/go-fdo-tool
ENTRYPOINT ["go-fdo-server"]
