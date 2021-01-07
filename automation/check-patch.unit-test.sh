#!/bin/bash -e

# This script should be able to execute functional tests against Kubernetes
# cluster on any environment with basic dependencies listed in
# check-patch.packages installed and docker running.

teardown() {
    cp $(find . -name "*junit*.xml") $ARTIFACTS
}

main() {
    source automation/check-patch.setup.sh
    cd ${TMP_PROJECT_PATH}
    trap teardown EXIT SIGINT SIGTERM SIGSTOP
    make bump-all
    make check
    make docker-build
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
