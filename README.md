# About

My [fido-device-onboard go-fdo-server](https://github.com/fido-device-onboard/go-fdo-server) Playground.

**NB** This targets [go-fdo-server 9469c980914f3fdb5ab86b10620e852ea2951e8a (2026-10-08T12:16:37.000Z)](https://github.com/fido-device-onboard/go-fdo-server/tree/9469c980914f3fdb5ab86b10620e852ea2951e8a).

# Usage

This can be tested [in docker compose](#usage-docker-compose).

## Usage (docker compose)

Install docker and docker compose.

Install the [`debian-13-uefi-amd64` vagrant box](https://github.com/rgl/debian-vagrant).

Recreate the playground:

```bash
docker compose down --volumes --remove-orphans --timeout=0
docker compose up --wait --build --detach
```

Initialize the Rendezvous service:

```bash
docker compose exec rendezvous \
    curl \
    --silent \
    --show-error \
    --request POST \
    http://rendezvous:8041/api/v2/device-ca \
    --header "Content-Type: application/x-pem-file" \
    --data-binary @/etc/ssl/private/device-ca-crt.pem
docker compose exec rendezvous \
    curl \
    --silent \
    --show-error \
    --request GET \
    http://rendezvous:8041/api/v2/device-ca \
    | jq
```

Initialize the Manufacturer service:

```bash
docker compose exec manufacturer \
    curl \
    --silent \
    --show-error \
    -X PUT  \
    -H 'Content-Type: application/json' \
    -d '[[{"protocol":"http"},{"dns":"rendezvous"},{"device_port":8041},{"owner_port":8041}]]' \
    http://manufacturer:8038/api/v2/rvinfo
docker compose exec manufacturer \
    curl \
    --silent \
    --show-error \
    -X GET \
    http://manufacturer:8038/api/v2/rvinfo \
    | jq
```

Initialize the Owner service:

```bash
docker compose exec owner \
    curl \
    --silent \
    --show-error \
    --request PUT \
    http://owner:8043/api/v2/rvto2addr \
    --header 'Content-Type: application/json' \
    --data-raw '[{"protocol":"http","dns":"owner","port":8043}]'
docker compose exec owner \
    curl \
    --silent \
    --show-error \
    --request GET \
    http://owner:8043/api/v2/rvto2addr \
    | jq
```

Do the Device Initialization (DI) protocol:

```bash
# NB the --device-info value will be included in the Device Ownership Voucher.
# NB add the --debug cli flag to debug.
# NB when the machine does not have a TPM, --blob /device/credential.blob could
#    be used to store the device credential in a plaintext file.
docker compose exec device ssh vagrant@device sudo \
    go-fdo-client \
    device-init \
    http://manufacturer:8038 \
    --key ec256 \
    --tpm /dev/tpmrm0 \
    --device-info '{"exampleKey":"exampleValue"}'
DEVICE_GUID="$(docker compose exec device ssh vagrant@device sudo \
    go-fdo-client \
    print \
    --tpm /dev/tpmrm0 \
    | perl -ne 'print "$1\n" if /^\s*GUID\s+([0-9a-f]+)$/i')"
echo "Device GUID: $DEVICE_GUID"
```

Show information about the [public Device Credential structure (aka DCTPM)](https://github.com/fido-device-onboard/go-fdo/blob/main/tpm/credential.go) stored on the TPM non-volatile (NV) storage:

```bash
docker compose exec device ssh vagrant@device sudo \
    tpm2 getcap handles-nv-index
docker compose exec device ssh vagrant@device sudo \
    tpm2 nvreadpublic 0x01D10001
```

Upload the secret Device Ownership Voucher to the Owner service:

**NB** Upon upload, the Owner will start the Transfer Ownership Zero (TO0)
protocol, to transfer the ownership of the device to itself. It must finish
successfully before the device can be onboarded.

```bash
DEVICE_VOUCHER="$(docker compose exec manufacturer \
    curl \
        --silent \
        --show-error \
        --request GET \
        --header 'Accept: application/x-pem-file' \
        "http://manufacturer:8038/api/v2/vouchers/$DEVICE_GUID")"
docker compose exec manufacturer \
    curl \
        --silent \
        --show-error \
        --request GET \
        --header 'Accept: application/json' \
        "http://manufacturer:8038/api/v2/vouchers/$DEVICE_GUID" \
        | jq
docker compose exec --no-tty manufacturer \
    go-fdo-tool \
    voucher \
    print \
    /dev/stdin \
    <<<"$DEVICE_VOUCHER"
docker compose exec manufacturer \
    curl \
        --silent \
        --show-error \
        --request POST \
        --header 'Content-Type: application/x-pem-file' \
        http://owner:8043/api/v2/vouchers \
        --data-binary "$DEVICE_VOUCHER"
# observe the logs until TO0 completed successfully.
# NB the logs should end with something like:
#       [22:23:24] INFO: successfully registered 'RVTO2Addr' to rendezvous server
#         url: http://rendezvous:8041
#         duration: 5m0s
#       [22:23:24] DEBUG: to0 scheduler: register 'RV2TO0Addr' completed
#         guid: 0d9f789605e33f74947eef1b15c029cc
#         refresh: 300
docker compose logs owner --follow
```

Prepare the device to be onboarded:

```bash
# generate a unique device service admin password.
# NB this will be uploaded to the owner service as one of the last steps of the
#    onboard process, by the fdo.upload FDO ServiceInfo Module (fsim), as
#    declared in the owner.yaml file.
docker compose exec --no-tty device ssh vagrant@device sudo bash -eux - <<EOF
install -m 700 -d /device
umask 077
openssl rand -out /device/admin-password.txt -hex 16
EOF
```

Onboard the device:

```bash
# NB add the --debug cli flag to debug.
docker compose exec device ssh vagrant@device sudo \
    go-fdo-client \
    onboard \
    --key ec256 \
    --kex ECDH256 \
    --tpm /dev/tpmrm0
docker compose exec device ssh vagrant@device sudo \
    go-fdo-client \
    print \
    --tpm /dev/tpmrm0
```

Show the device state from the Owner service viewpoint:

```bash
DEVICE_OLD_GUID="$DEVICE_GUID"
DEVICE_INFO="$(docker compose exec owner \
    curl \
        --silent \
        --show-error \
        --request GET \
        http://owner:8043/api/v2/devices \
        --url-query "old_guid=$DEVICE_OLD_GUID")"
DEVICE_GUID="$(jq -r '.devices[0].guid' <<<"$DEVICE_INFO")"
jq <<<"$DEVICE_INFO"
```

Access the Device Admin Page from the Owner service viewpoint:

```bash
DEVICE_ADMIN_PASSWORD="$(docker compose exec owner \
    cat "/var/run/fdo-owner/device-files/$DEVICE_GUID/admin-password.txt")"
docker compose exec owner \
    curl \
    --silent \
    --show-error \
    --request GET \
    --user "admin:$DEVICE_ADMIN_PASSWORD" \
    http://device:8080/admin
```

Poke the FDO services database:

```bash
docker compose exec rendezvous sqlite3 /var/run/fdo-rendezvous/rendezvous.sqlite '.schema'
docker compose exec rendezvous sqlite3 /var/run/fdo-rendezvous/rendezvous.sqlite '.dump'
docker compose exec manufacturer sqlite3 /var/run/fdo-manufacturer/manufacturer.sqlite '.schema'
docker compose exec manufacturer sqlite3 /var/run/fdo-manufacturer/manufacturer.sqlite '.dump'
docker compose exec manufacturer sqlite3 /var/run/fdo-manufacturer/manufacturer.sqlite 'select * from vouchers'
docker compose exec owner sqlite3 /var/run/fdo-owner/owner.sqlite '.schema'
docker compose exec owner sqlite3 /var/run/fdo-owner/owner.sqlite '.dump'
docker compose exec owner sqlite3 /var/run/fdo-owner/owner.sqlite 'select * from vouchers'
```

Poke the FDO services OpenAPI definitions:

```bash
docker compose exec rendezvous curl --silent http://rendezvous:8041/api/openapi.json | jq
docker compose exec manufacturer curl --silent http://manufacturer:8038/api/openapi.json | jq
docker compose exec owner curl --silent http://owner:8043/api/openapi.json | jq
xdg-open http://localhost:8041/api/openapi.json | jq    # rendezvous
xdg-open http://localhost:8041/api/docs/                # rendezvous
xdg-open http://localhost:8038/api/openapi.json | jq    # manufacturer
xdg-open http://localhost:8038/api/docs/                # manufacturer
xdg-open http://localhost:8043/api/openapi.json | jq    # owner
xdg-open http://localhost:8043/api/docs/                # owner
```

Destroy the playground:

```bash
docker compose down --volumes --remove-orphans --timeout=0
```

## References

* [What is FIDO Device Onboard (FDO)?](https://fidoalliance.org/device-onboarding-overview/).
* [FIDO Device Onboarding Resources](https://fidoalliance.org/device-onboarding-overview/device-onboarding-resources/).
* [FDO Specifications](https://fidoalliance.org/specifications/download-iot-specifications/)
* https://github.com/fido-device-onboard/go-fdo-server/blob/main/docs/user-guide/quick-start.md
* https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/10/html/composing_installing_and_managing_rhel_for_edge_images/automatically-provisioning-and-onboarding-rhel-for-edge-devices-with-fdo
* [FIDO Workshop Part 1: Introductions And How FDO Works](https://www.youtube.com/watch?v=JL_C9vSlY_M&list=PLKZSArYQptsP14GmI07Ssp3OjtBmTo1dv)
