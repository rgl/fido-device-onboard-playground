#!/usr/bin/env bash
set -Eeuo pipefail

function buildCloudInitData {
  local dest="$1"
  local cidata="${dest%.img}-cidata"
  local key="$HOME/.ssh/id_ed25519"

  ssh-keygen -t ed25519 -N "" -f "$key"

  cat >~/.ssh/config <<EOF
Host $(hostname)
  StrictHostKeyChecking no
EOF

  install -d "$cidata"

  cat >"$cidata/user-data" <<EOF
#cloud-config
users:
  - name: vagrant
    lock_passwd: false
    ssh_authorized_keys:
      - '$(cat "$key.pub")'
mounts:
  - [files, /mnt/files, 9p, "trans=virtio,ro,nofail", "0", "0"]
apt:
  conf: |
    APT::Install-Recommends "false";
package_update: true
packages:
  - curl
  - tpm2-tools
write_files:
  - path: /etc/systemd/system/device.service
    permissions: '0644'
    owner: root:root
    content: |
      [Unit]
      Description=Device
      After=network.target

      [Service]
      Type=simple
      ExecStart=/usr/bin/device
      Restart=on-failure

      [Install]
      WantedBy=multi-user.target
runcmd:
  - [install, /mnt/files/go-fdo-client, /usr/bin/go-fdo-client]
  - [install, /mnt/files/device, /usr/bin/device]
  - [systemctl, daemon-reload]
  - [systemctl, enable, device.service]
  - [systemctl, start, device.service]
EOF

  cat >"$cidata/meta-data" <<EOF
EOF

  xorriso \
    -as genisoimage \
    -output "$cidata.iso" \
    -volid cidata \
    -joliet \
    -rock \
    "$cidata"

  xorriso \
    -dev "$cidata.iso" \
    -find / \
    -type f

  ARGUMENTS+=" -drive file=$cidata.iso,media=cdrom,readonly=on"

  ARGUMENTS+=" -fsdev local,id=fsdev1,path=/files,security_model=none,readonly=on"
  ARGUMENTS+=" -device virtio-9p-pci,id=fs1,fsdev=fsdev1,mount_tag=files"

  ARGUMENTS="${ARGUMENTS# }"
}

function downloadRetry {
  if (( $# < 6 )); then
    error "downloadRetry requires a destination, connection count, delay, description, a minimum size, and url."
    return 2
  fi

  local dest="$1"
  local connections="$2"
  local seconds="$3"
  local description="$4"
  local minimum="$5"
  local url="$6"
  if [[ "$url" != file://* ]]; then
    echo "ERROR: URL must start with file://" >&2
    return 1
  fi
  local src="${url#file://}"

  buildCloudInitData "$dest"

  qemu-img convert -f qcow2 "$src" -O raw "$dest"
}
