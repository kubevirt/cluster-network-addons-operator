#!/bin/bash -xe

# This script should be able to execute functional tests against OCP cluster on
# any environment with basic dependencies listed in check-patch.packages
# installed and docker running.
#
# yum -y install automation/check-patch.packages
# automation/check-patch.e2e-lifecycle-ocp.sh

teardown() {
    make cluster-down
}

main() {
    export KUBEVIRT_PROVIDER='ocp-4.3'

    source automation/check-patch.setup.sh
    cd ${TMP_PROJECT_PATH}

    if versionChanged; then
        # Since we cannot test upgrade of to-be-released version, drop it from the lifecycle tests
        to_be_released=$(hack/version.sh)
        export RELEASES_DESELECTOR="${to_be_released}"
    else
        # Don't run all upgrade tests in regular PRs, stick to those released under HCO
        export RELEASES_SELECTOR="{0.18.0,0.23.0,0.27.0,99.0.0}"
    fi

    make cluster-down
    make cluster-up
    trap teardown EXIT SIGINT SIGTERM SIGSTOP
    make cluster-operator-push
    make test/e2e/lifecycle
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
