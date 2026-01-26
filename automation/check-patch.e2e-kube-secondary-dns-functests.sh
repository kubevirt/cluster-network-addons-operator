#!/usr/bin/env bash

set -xeuE

# This script should be able to execute kube secondary dns
# functional tests against Kubernetes cluster with
# CNAO built with latest changes, on any
# environment with basic dependencies listed in
# check-patch.packages installed and docker running.
#
# yum -y install automation/check-patch.packages
# automation/check-patch.e2e-kube-secondary-dns-functests.sh

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

    export KUBEVIRT_NUM_NODES=3
    # Spin-up ephemeral cluster with latest CNAO
    # this script also exports KUBECONFIG, and fetch $COMPONENT repository
    export CNAO_DEPLOY_KUBEVIRT=true
    COMPONENT="kube-secondary-dns" source automation/components-functests.setup.sh

    trap teardown EXIT

    echo "check cross-node connectivity before network-policy"
    ./automation/check-pod-to-pod-ping.sh

    echo "Simulate network restrictions on CNAO namespace"
    ./hack/install-network-policy.sh

    echo "check cross-node connectivity after network-policy"
    ./automation/check-pod-to-pod-ping.sh

    cd ${TMP_COMPONENT_PATH}
    make create-nodeport
    echo "Run kube-secondary-dns functional tests"
    make functest
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
