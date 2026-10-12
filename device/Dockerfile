# syntax=docker.io/docker/dockerfile:1.28

# see https://github.com/fido-device-onboard/go-fdo-client
# see https://github.com/fido-device-onboard/go-fdo-client/tree/0ccf734fefdb641e1dc5f3ccc9ad1e408edf71bd # 2026-07-28T15:49:42.000Z
FROM golang:1.27.2-trixie AS build-fdo-client
WORKDIR /src
RUN curl -L https://github.com/fido-device-onboard/go-fdo-client/archive/0ccf734fefdb641e1dc5f3ccc9ad1e408edf71bd.tar.gz | tar xz --strip-components=1
RUN make build

FROM golang:1.27.2-trixie AS build
WORKDIR /src
COPY go.* ./
RUN go mod download
COPY *.go ./
RUN go build -o device -ldflags="-s -w" -trimpath

# see https://github.com/qemus/qemu
FROM qemux/qemu:7.50
RUN <<RUN_EOF
#!/usr/bin/bash
set -euxo pipefail
apt-get update
apt-get -y install --no-install-recommends \
    openssh-client
apt-get clean
rm -rf /var/lib/apt/lists/*
RUN_EOF
RUN <<RUN_EOF
#!/usr/bin/bash
set -euxo pipefail
sed -i -E 's,http://\*.+,file://* ) ;;,g' /run/install.sh
RUN_EOF
COPY download.sh /run/download.sh
COPY --from=build-fdo-client /src/go-fdo-client /files/go-fdo-client
COPY --from=build /src/device /files/device
