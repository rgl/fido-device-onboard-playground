# syntax=docker.io/docker/dockerfile:1.28

FROM debian:trixie-slim
RUN <<EOF
#!/bin/bash
set -euxo pipefail
apt-get update
apt-get install -y --no-install-recommends \
  ca-certificates \
  openssl 
rm -rf /var/lib/apt/lists/*
EOF
COPY --chmod=755 create.sh /etc/ssl/private/create.sh
CMD ["/etc/ssl/private/create.sh"]
