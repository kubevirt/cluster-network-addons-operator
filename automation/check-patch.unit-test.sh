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
    make bump-all
    make check
    verify_metrics_docs_updated
    make lint-metrics
    make lint-monitoring
    make docker-build
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
