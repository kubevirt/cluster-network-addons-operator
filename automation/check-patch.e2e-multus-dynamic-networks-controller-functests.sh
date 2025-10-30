#!/usr/bin/env bash

set -xeuE

# This script should be able to execute macvtap-cni
# functional tests against Kubernetes cluster with
# CNAO built with latest changes, on any
# environment with basic dependencies listed in
# check-patch.packages installed and docker running.
#
# yum -y install automation/check-patch.packages
# automation/check-patch.e2e-macvtap-functests.sh

teardown() {
    cp $(find . -name "*junit*.xml") $ARTIFACTS || true
    rm -rf "${TMP_COMPONENT_PATH}"
    cd ${TMP_PROJECT_PATH}
    make cluster-down
}

main() {
    # Setup CNAO and artifacts temp directory
    source automation/check-patch.setup.sh
    cd ${TMP_PROJECT_PATH}

    # Spin-up ephemeral cluster with latest CNAO
    # this script also exports KUBECONFIG, and fetch $COMPONENT repository
    COMPONENT="multus-dynamic-networks" source automation/components-functests.setup.sh

    trap teardown EXIT

    echo "Install golang ..."
    cp hack/install-go.sh ${TMP_COMPONENT_PATH}/hack/install-go.sh
    cp hack/go-version.sh ${TMP_COMPONENT_PATH}/hack/go-version.sh
    cd ${TMP_COMPONENT_PATH}
    BIN_DIR="${TMP_COMPONENT_PATH}/build/_output/go1.18/bin/"
    mkdir -p "$BIN_DIR"
    hack/install-go.sh "$BIN_DIR"
    export GOROOT="$BIN_DIR/go/"
    export GOBIN="$GOROOT/bin/"
    export PATH="$GOBIN:$PATH"
    go version

    echo "Run multus-dynamic-networks functional tests"
    export LOWER_DEVICE="eth0"
    make e2e/test
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
