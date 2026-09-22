# syntax=docker.io/docker/dockerfile:1.27

# see https://github.com/fido-device-onboard/go-fdo-server
# see https://github.com/fido-device-onboard/go-fdo-server/tree/a1dfe4cf17a8a4736bf4bd56644b900bc054a150 # 2026-08-31T14:57:14.000Z
FROM golang:1.27.1-trixie AS build
RUN <<RUN_EOF
#!/bin/bash
set -euxo pipefail
curl --fail --show-error https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key >/etc/apt/trusted.gpg.d/deb.nodesource.com.gpg
# see https://wiki.debian.org/DebianRepository/Format
# see node 24.x: https://deb.nodesource.com/node_24.x/dists/nodistro/Release
# see node 24.x: https://deb.nodesource.com/node_24.x/dists/nodistro/main/binary-amd64/Packages
cat >/etc/apt/sources.list.d/deb.nodesource.com.sources <<EOF
Types: deb
URIs: https://deb.nodesource.com/node_24.x
Suites: nodistro
Components: main
Signed-By: /etc/apt/trusted.gpg.d/deb.nodesource.com.gpg
EOF
apt-get update
apt-get install -y --no-install-recommends \
    nodejs=24.21.0-*
RUN_EOF
WORKDIR /src
RUN curl -L https://github.com/fido-device-onboard/go-fdo-server/archive/a1dfe4cf17a8a4736bf4bd56644b900bc054a150.tar.gz | tar xz --strip-components=1
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
