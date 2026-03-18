#!/bin/bash -xe

# This script should be able to execute workflow functional tests against
# Kubernetes cluster on any environment with basic dependencies listed in
# check-patch.packages installed and docker running.
#
# yum -y install automation/check-patch.packages
# automation/check-patch.e2e-workflow-k8s.sh

teardown() {
    # Don't fail if there is no logs
    # cp ${E2E_LOGS}/workflow/*.log ${ARTIFACTS} || true
    cp ${E2E_LOGS}/compliance/*.log ${ARTIFACTS} || true

    make cluster-down
}

main() {
    source automation/check-patch.setup.sh
    cd ${TMP_PROJECT_PATH}

    make cluster-down
    make cluster-up
    trap teardown EXIT SIGINT SIGTERM SIGSTOP

    ./hack/deploy-kubevirt.sh
    make cluster-operator-push
    make cluster-operator-install

    echo "Simulate network restriction on CNAO namespace"
    ./hack/install-network-policy.sh

    # make E2E_TEST_EXTRA_ARGS="-ginkgo.noColor --ginkgo.junit-report=$ARTIFACTS/junit.functest.xml" test/e2e/workflow

    ./hack/install-tls-compliance-operator.sh

    (
        export KUBECONFIG="$(./_kubevirtci/cluster-up/kubeconfig.sh)"
        export OCI_BIN="podman"
        export TLSREPORT_IMAGE="quay.io/omergi0/kubectl-tlsreport:v0.0.10"
        export ARTIFACTS
        make E2E_TEST_EXTRA_ARGS="-ginkgo.no-color --ginkgo.junit-report=$ARTIFACTS/junit.functest.xml -ginkgo.v -ginkgo.progress" test/e2e/compliance
    )
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
