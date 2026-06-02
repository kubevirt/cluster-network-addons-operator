#!/bin/bash -e

# This script should be able to execute functional tests against Kubernetes
# cluster on any environment with basic dependencies listed in
# check-patch.packages installed and docker running.

verify_metrics_docs_updated() {
  make generate-doc
  git difftool -y --trust-exit-code --extcmd=./hack/diff-csv.sh
}

main() {
    source automation/check-patch.setup.sh
    cd ${TMP_PROJECT_PATH}

    # Skip bump-all in unit tests to avoid expensive multi-platform image builds
    # Component bump validation should be done in a separate build/integration job
    # See: https://github.com/kubevirt/cluster-network-addons-operator/issues/2732

    make vendor
    if ! make check; then
        echo "error: Uncommitted changes found after vendor check. \
        Make sure go.mod is up to date and all 'make vendor' output is commited"
    fi

    verify_metrics_docs_updated
    make lint-metrics
    make lint-monitoring

    # Build only for current architecture to speed up CI
    # Multi-platform builds should be done in release/integration jobs
    export PLATFORMS=$(uname -m | sed 's/x86_64/linux\/amd64/;s/aarch64/linux\/arm64/;s/s390x/linux\/s390x/')
    make docker-build
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
