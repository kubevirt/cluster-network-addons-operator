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
    rm -rf "${TMP_COMPONENT_PATH}"
    cp $(find . -name "*junit*.xml") $ARTIFACTS || true
    make cluster-down
}

main() {
    # Setup CNAO and artifacts temp directory
    source automation/check-patch.setup.sh
    cd ${TMP_PROJECT_PATH}

    # Spin-up ephemeral cluster with latest CNAO
    # this script also exports KUBECONFIG, and fetch $COMPONENT repository
    COMPONENT="macvtap-cni" source automation/components-functests.setup.sh

    trap teardown EXIT

    echo "Run macvtap-cni functional tests"
    cd ${TMP_COMPONENT_PATH}

    make test/e2e
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
