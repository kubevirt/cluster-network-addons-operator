#!/usr/bin/env bash

set -xeu

# This script should be able to execute bridge-marker
# functional tests against Kubernetes cluster with
# CNAO built with latest changes, on any
# environment with basic dependencies listed in
# check-patch.packages installed and docker running.
#
# yum -y install automation/check-patch.packages
# automation/check-patch.e2e-bridge-marker-functests.sh

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
    COMPONENT="bridge-marker" source automation/components-functests.setup.sh

    trap teardown EXIT

    echo "Run bridge-marker functional tests"
    cd ${TMP_COMPONENT_PATH}

    if ! KUBECONFIG=$KUBECONFIG FUNC_TEST_ARGS="--ginkgo.noColor --junit-output=$ARTIFACTS/junit.functest.xml" make functest; then
        ./cluster/kubectl.sh logs -n kube-system -l app=bridge-marker
        return 1
    fi
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
