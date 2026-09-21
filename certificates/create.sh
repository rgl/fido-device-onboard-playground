#!/usr/bin/bash
set -euxo pipefail

function create-self-signed-certificate {
    local name="$1"
    local subject="$2"
    local -a dn=(); IFS='/' read -r -a dn <<<"${subject#/}"
    cat >"$name-crt.conf" <<EOF
[req]
prompt = no
days = 365
distinguished_name = dn
x509_extensions = extensions
default_md = sha256
[dn]
$(printf '%s\n' "${dn[@]}")
[extensions]
EOF
    if [[ "$name" == *-ca ]]; then
        cat >>"$name-crt.conf" <<EOF
basicConstraints = critical,CA:TRUE
keyUsage = critical,keyCertSign,cRLSign
subjectKeyIdentifier = hash
EOF
    fi
    openssl ecparam \
        -name prime256v1 \
        -genkey \
        -out "$name-key.der" \
        -outform der
    openssl req \
        -config "$name-crt.conf" \
        -x509 \
        -key "$name-key.der" \
        -keyform der \
        -out "$name-crt.pem" \
        -outform pem
    openssl x509 \
        -noout \
        -text \
        -in "$name-crt.pem"
}

create-self-signed-certificate manufacturer '/O=Example/OU=FDO/CN=Manufacturer'
create-self-signed-certificate device-ca    '/O=Example/OU=FDO/CN=Device CA'
create-self-signed-certificate owner        '/O=Example/OU=FDO/CN=Owner'
