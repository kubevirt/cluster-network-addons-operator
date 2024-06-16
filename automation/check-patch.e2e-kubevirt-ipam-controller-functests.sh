#!/usr/bin/env bash

set -xeuE

# automation/check-patch.e2e-kubevirt-ipam-controller-functests.sh

teardown() {
    cd ${TMP_COMPONENT_PATH}
    make cluster-down || true
    rm -rf "${TMP_COMPONENT_PATH}"
}

main() {
    if [ "$GITHUB_ACTIONS" == "true" ]; then
        ARCH="amd64"
        OS_TYPE="linux"
        kubevirt_version="$(curl -L https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirt/stable.txt)"
        kubevirt_release_url="https://github.com/kubevirt/kubevirt/releases/download/${kubevirt_version}"
        cli_name="virtctl-${kubevirt_version}-${OS_TYPE}-${ARCH}"
        curl -LO "${kubevirt_release_url}/${cli_name}"
        mv ${cli_name} virtctl
        chmod +x virtctl
        mv virtctl /usr/local/bin
    fi

    # Setup CNAO and artifacts temp directory
    source automation/check-patch.setup.sh
    cd ${TMP_PROJECT_PATH}

    export USE_KUBEVIRTCI=false
    COMPONENT="kubevirt-ipam-controller" source automation/components-functests.setup.sh

    cd ${TMP_COMPONENT_PATH}
    export KIND_ARGS="-ic -i6 -mne"
    make cluster-up
    export KUBECONFIG=${TMP_COMPONENT_PATH}/.output/kubeconfig

    trap teardown EXIT

    cd ${TMP_PROJECT_PATH}
    export KUBEVIRT_PROVIDER=external
    export DEV_IMAGE_REGISTRY=localhost:5000
    deploy_cnao
    deploy_cnao_cr
    ./hack/deploy-kubevirt.sh

    cd ${TMP_COMPONENT_PATH}
    echo "Run kubevirt-ipam-controller functional tests"
    make test-e2e
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
