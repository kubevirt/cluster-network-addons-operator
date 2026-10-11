#!/usr/bin/env bash

set -euo pipefail

if [[ "${1:-}" != "exec" ]]; then
    echo "unsupported ovs-cni OCI operation: ${1:-}" >&2
    exit 1
fi
shift

node="${1:?missing node name}"
shift

project_path="${TMP_PROJECT_PATH:?missing CNAO project path}"
(
    cd "${project_path}"
    source cluster/cluster.sh
    exec "${CLUSTER_PATH}/cluster-up/ssh.sh" "${node}" -- sudo "$@"
)
