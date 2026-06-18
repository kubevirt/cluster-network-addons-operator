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

    # Build only for current architecture to speed up CI
    # Multi-platform builds (amd64, s390x, arm64) should be done in release/integration jobs
    # This reduces CI time from ~2.5+ minutes to under 1 minute by avoiding QEMU emulation
    export PLATFORMS="linux/$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')"

    make bump-all
    if ! make check; then
        echo "error: Uncommitted changes found after bump check. \
        If you are performing a component bump, recheck all created manifests and image URL are valid. \
        If you are not attempting a component bump, please contact the repo's maintainer for further analysis"
    fi

    make vendor
    if ! make check; then
        echo "error: Uncommitted changes found after vendor check. \
        Make sure go.mod is up to date and all 'make vendor' output is commited"
    fi

    verify_metrics_docs_updated
    make lint-metrics
    make lint-monitoring
    make docker-build
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
