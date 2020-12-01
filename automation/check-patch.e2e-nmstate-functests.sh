#!/usr/bin/env bash

set -xeu

# This script should be able to execute Kubernetes-nmstate
# functional tests against Kubernetes cluster with
# CNAO built with latest changes, on any
# environment with basic dependencies listed in
# check-patch.packages installed and docker running.
#
# yum -y install automation/check-patch.packages
# automation/check-patch.e2e-nmstate-functests.sh

teardown() {
    rm -rf "${TMP_COMPONENT_PATH}"
    make cluster-down
}

main() {
    # Setup CNAO and artifacts temp directory
    source automation/check-patch.setup.sh
    cd ${TMP_PROJECT_PATH}

    # Spin-up ephemeral cluster with latest CNAO
    # this script also exports KUBECONFIG, and fetch $COMPONENT repository
    export KUBEVIRT_NUM_NODES=3 # 1 master, 2 workers
    export KUBEVIRT_NUM_SECONDARY_NICS=2
    COMPONENT="nmstate" source automation/components-functests.setup.sh

    trap teardown EXIT

    echo "Configure test parameters"
    export TIMEOUT=1h
    export NAMESPACE=cluster-network-addons
    export KUBECTL=${TMP_PROJECT_PATH}/cluster/kubectl.sh
    export SSH=${TMP_PROJECT_PATH}/cluster/ssh.sh
    export CLUSTER_PATH=${TMP_PROJECT_PATH}/_kubevirtci/

    echo "Run nmstate functional tests"
    cd ${TMP_COMPONENT_PATH}

    make test-e2e-handler \
        E2E_TEST_TIMEOUT=$TIMEOUT \
        e2e_test_args="-noColor" \
        E2E_TEST_SUITE_ARGS="--junit-output=$ARTIFACTS/junit.functest.xml" \
        OPERATOR_NAMESPACE=$NAMESPACE \
        CLUSTER_PATH=$CLUSTER_PATH \
        KUBECONFIG=$KUBECONFIG \
        KUBECTL=$KUBECTL
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
