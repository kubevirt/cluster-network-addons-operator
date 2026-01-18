#!/bin/bash -xe

# This script should be able to execute KubeMacPool
# functional tests against Kubernetes cluster with
# CNAO built with latest changes, on any
# environment with basic dependencies listed in
# check-patch.packages installed and docker running.
#
# yum -y install automation/check-patch.packages
# automation/check-patch.e2e-kubemacpool-functests.sh


teardown() {
    # Copy kubemacpool failure logs to CNAO artifacts before cleanup
    cp -r ${TMP_COMPONENT_PATH}/tests/_out/* $ARTIFACTS || true
    rm -rf "${TMP_COMPONENT_PATH}"
    cd ${TMP_PROJECT_PATH}
    make cluster-down
}

main() {
    # Setup CNAO and artifacts temp directory
    source automation/check-patch.setup.sh
    cd ${TMP_PROJECT_PATH}

    export KUBEVIRT_NUM_NODES=${KUBEVIRT_NUM_NODES:-3}
    # Spin-up ephemeral cluster with latest CNAO
    # this script also exports KUBECONFIG, and fetch $COMPONENT repository
    COMPONENT="kubemacpool" source automation/components-functests.setup.sh

    trap teardown EXIT

    echo "Simulate network restrictions on CNAO namespace"
    ./hack/install-network-policy.sh

    echo "Deploy KubeVirt latest stable release"
    ./hack/deploy-kubevirt.sh

    # Run KubeMacPool functional tests
    cd ${TMP_COMPONENT_PATH}

    export CLUSTER_ROOT_DIRECTORY=${TMP_PROJECT_PATH}
    KUBECONFIG=${KUBECONFIG} make E2E_TEST_EXTRA_ARGS="-ginkgo.noColor -test.outputdir=$ARTIFACTS --ginkgo.junit-report=$ARTIFACTS/junit.functest.xml --ginkgo.label-filter=!vm-opt-in&&!mac-range-day2-update" functest
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
