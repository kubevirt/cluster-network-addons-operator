#!/bin/bash -xe

# This script should be able to execute lifecycle functional tests against
# Kubernetes cluster on any environment with basic dependencies listed in
# check-patch.packages installed and docker running.
#
# yum -y install automation/check-patch.packages
# automation/check-patch.e2e-lifecycle-k8s.sh

teardown() {
    # Don't fail if there is no logs
    cp ${E2E_LOGS}/lifecycle/*.log ${ARTIFACTS} || true
    make cluster-down
}

versionChanged() {
    git diff --name-only HEAD HEAD~1 | grep version/version.go
}

main() {
    export KUBEVIRT_PROVIDER='k8s-1.25'

    source automation/check-patch.setup.sh
    cd ${TMP_PROJECT_PATH}

    if versionChanged; then
        # Since we cannot test upgrade of to-be-released version, drop it from the lifecycle tests
        to_be_released=$(hack/version.sh)
        export RELEASES_DESELECTOR="${to_be_released}"
        export E2E_TEST_TIMEOUT=4h
    else
        # Don't run all upgrade tests in regular PRs, stick to those released under HCO
        export RELEASES_SELECTOR="{0.65.10,0.76.3,0.79.1,0.85.0,99.0.0}"
    fi

    make cluster-down
    make cluster-up
    trap teardown EXIT SIGINT SIGTERM SIGSTOP
    make cluster-operator-push
    make E2E_TEST_EXTRA_ARGS="-ginkgo.noColor --ginkgo.junit-report=$ARTIFACTS/junit.functest.xml" test/e2e/lifecycle
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
