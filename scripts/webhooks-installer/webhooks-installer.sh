#!/usr/bin/env bash

function patch_webhook() {
  OLD_WEBHOOK=$(kubectl get "$WEBHOOK_TYPE" "$WEBHOOK_NAME" -o yaml)

  NEW_WEBHOOK=${OLD_WEBHOOK//caBundle: eHh4Cg==/caBundle: "$CA_BASE64"}
  echo "$NEW_WEBHOOK" | kubectl apply -f -
}

function create_certificate() {

  tmpdir="/tmp/liqo/$SECRET_NAME/ssl"
  mkdir -p $tmpdir

  cat <<EOF >> "${tmpdir}"/csr.conf

    [req]
    req_extensions = v3_req
    distinguished_name = req_distinguished_name
    [req_distinguished_name]
    [ v3_req ]
    basicConstraints = CA:FALSE
    keyUsage = nonRepudiation, digitalSignature, keyEncipherment
    extendedKeyUsage = serverAuth
    subjectAltName = @alt_names
    [alt_names]
    DNS.1 = ${SERVICE_NAME}
    DNS.2 = ${SERVICE_NAME}.${NAMESPACE}
    DNS.3 = ${SERVICE_NAME}.${NAMESPACE}.svc

EOF

  CSR_NAME=${SERVICE_NAME}.${NAMESPACE}

  # Csr creation
  openssl genrsa -out "$tmpdir/server-key.pem" 2048
  openssl req -new -key "$tmpdir/server-key.pem" -subj "/CN=${SERVICE_NAME}.${NAMESPACE}.svc" -out "$tmpdir/$CSR_NAME".csr -config "$tmpdir/csr.conf"

  # self-signed certificate creation
  openssl genrsa -out "$tmpdir/myCA.key" 2048
  openssl req -x509 -new -nodes -key "$tmpdir/myCA.key" -sha256 -days 1825 -out "$tmpdir/myCA.pem" -subj "/C=GB/ST=London/L=London/O=Global Security/OU=IT Department/CN=example.com"

  # certificate signature
  openssl x509 -req -in "$tmpdir/$CSR_NAME.csr" -CA "$tmpdir/myCA.pem" -CAkey "$tmpdir/myCA.key" -CAcreateserial -out "$tmpdir/$CSR_NAME.crt" -days 825 -sha256

  CA_BASE64=$(base64 "$tmpdir/myCA.pem" | tr -d '\n')

    # create the secret with CA cert and server cert/key
  kubectl create secret generic "$SECRET_NAME" \
        --from-file=key.pem="$tmpdir/server-key.pem" \
        --dry-run=client -o yaml | kubectl -n "$NAMESPACE" apply -f -
}

if [ $# -ne 2 ]; then
  echo "illegal number of parameters"
  exit 1
fi

INPUT_DIR=$1
NAMESPACE=$2

if [ ! -x "$(command -v openssl)" ]; then
    echo "openssl not found"
    exit 1
fi

for WEBHOOK_PATH in "$INPUT_DIR"/*; do
  SERVICE_NAME=
  SECRET_NAME=
  WEBHOOK_TYPE=

  WEBHOOK_NAME=$(basename "$WEBHOOK_PATH")
  cmd="source $WEBHOOK_PATH/vars"
  eval "$cmd"
  if [ -z $SERVICE_NAME ] || [ -z $SECRET_NAME ] || [ -z $WEBHOOK_TYPE ]; then
    echo "data missing in configmap $WEBHOOK_PATH"
    continue
  fi
  create_certificate
  patch_webhook
done
