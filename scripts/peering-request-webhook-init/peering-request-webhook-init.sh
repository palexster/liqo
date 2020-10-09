#!/bin/bash

set -e

usage() {
    cat <<EOF
Create the admission webhook deputed to chack new PeeringRequests
The following flags are optional.
       --input-env-file   The output directory for env variables
EOF
    exit 1
}

while [[ $# -gt 0 ]]; do
    case ${1} in
        --input-env-file)
            inputenvfile="$2"
            shift
            ;;
        *)
            usage
            ;;
    esac
    shift
done

[ -z "${inputenvfile}" ] && inputenvfile=/etc/environment/liqo/env

# shellcheck source=/dev/null
source ${inputenvfile}

CACRT=$(< /var/run/secrets/kubernetes.io/serviceaccount/ca.crt base64 | sed ':a;N;$!ba;s/\n//g')

# shellcheck disable=SC2154
cat <<EOF | kubectl apply -f -

EOF

exit 0