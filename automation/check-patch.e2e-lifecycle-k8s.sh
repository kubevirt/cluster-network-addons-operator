#!/bin/bash -xe

# This script should be able to execute lifecycle functional tests against
# Kubernetes cluster on any environment with basic dependencies listed in
# check-patch.packages installed and docker running.
#
# yum -y install automation/check-patch.packages
# automation/check-patch.e2e-lifecycle-k8s.sh

teardown() {
    make cluster-down
}

main() {
    export KUBEVIRT_PROVIDER='k8s-1.17.0'

    source automation/check-patch.e2e.setup.sh
    cd ${TMP_PROJECT_PATH}

    make cluster-down
    make cluster-up
    trap teardown EXIT SIGINT SIGTERM SIGSTOP
    make cluster-operator-push
    make test/e2e/lifecycle
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
