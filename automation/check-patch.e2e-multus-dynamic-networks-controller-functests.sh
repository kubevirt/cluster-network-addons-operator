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

    # Patch e2e test timeout for IPAM hot-unplug test (issue #2766)
    # Increase timeout from 15s to 30s to account for slower IPAM cleanup in CI
    echo "Patching e2e test timeout for hot-unplug operations..."
    if [ -f "e2e/e2e_test.go" ]; then
        # Replace timeout variable assignments from 15s to 30s
        # This handles patterns like: timeout := 15 * time.Second
        sed -i 's/timeout := 15 \* time\.Second/timeout := 30 * time.Second/g' e2e/e2e_test.go
        sed -i 's/timeout = 15 \* time\.Second/timeout = 30 * time.Second/g' e2e/e2e_test.go

        # Also handle direct usage in Eventually calls
        # Pattern: Eventually(..., 15*time.Second, ...)
        sed -i 's/\(Eventually([^,]*,\) 15\*time\.Second/\1 30*time.Second/g' e2e/e2e_test.go

        # Handle const timeout definitions
        sed -i 's/const timeout = 15 \* time\.Second/const timeout = 30 * time.Second/g' e2e/e2e_test.go

        echo "Timeout patch applied successfully"
    else
        echo "Warning: e2e/e2e_test.go not found, skipping timeout patch"
    fi

    export LOWER_DEVICE="eth0"
    make e2e/test
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
