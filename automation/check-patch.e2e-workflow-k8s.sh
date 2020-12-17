#!/bin/bash -xe

# This script should be able to execute workflow functional tests against
# Kubernetes cluster on any environment with basic dependencies listed in
# check-patch.packages installed and docker running.
#
# yum -y install automation/check-patch.packages
# automation/check-patch.e2e-workflow-k8s.sh

teardown() {
    make cluster-down
}

main() {
    export KUBEVIRT_PROVIDER='k8s-1.19'

    source automation/check-patch.setup.sh
    cd ${TMP_PROJECT_PATH}

    make cluster-down
    make cluster-up
    trap teardown EXIT SIGINT SIGTERM SIGSTOP
    make cluster-operator-push
    make cluster-operator-install
    make E2E_TEST_EXTRA_ARGS="-ginkgo.noColor --junit-output=$ARTIFACTS/junit.functest.xml" test/e2e/workflow
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
