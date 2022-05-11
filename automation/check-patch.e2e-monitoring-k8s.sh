#!/bin/bash -xe

# This script should be able to execute workflow functional tests against
# Kubernetes cluster on any environment with basic dependencies listed in
# check-patch.packages installed and docker running.
#
# yum -y install automation/check-patch.packages
# automation/check-patch.e2e-workflow-k8s.sh

teardown() {
    # Don't fail if there is no logs
    cp ${E2E_LOGS}/workflow/*.log ${ARTIFACTS} || true
    make cluster-down
}

main() {
    export KUBEVIRT_PROVIDER='k8s-1.23'

    source automation/check-patch.setup.sh
    cd ${TMP_PROJECT_PATH}

    export MONITORING_NAMESPACE="monitoring"
    make cluster-down
    make cluster-up
    trap teardown EXIT SIGINT SIGTERM
    make cluster-operator-push
    make cluster-operator-install
    make E2E_TEST_EXTRA_ARGS="-ginkgo.noColor --ginkgo.junit-report=$ARTIFACTS/junit.functest.xml" test/e2e/monitoring
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
