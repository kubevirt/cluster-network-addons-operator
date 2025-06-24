#!/usr/bin/env bash

set -xeuE

# automation/check-patch.e2e-kubevirt-ipam-controller-functests.sh

GITHUB_ACTIONS=${GITHUB_ACTIONS:-false}
LOGS_DIR=test/e2e/_output
WORK_DIR=$(pwd)

teardown() {
    cd ${WORK_DIR}
    rm -rf ${LOGS_DIR}
    mkdir -p ${LOGS_DIR}
    cp ${TMP_COMPONENT_PATH}/${LOGS_DIR}/*.log ${LOGS_DIR} || true

    cd ${TMP_COMPONENT_PATH}
    make cluster-down || true
    rm -rf "${TMP_COMPONENT_PATH}"
}

increase_ulimit() {
    if [ -z "${OCI_BIN}" ];then
      export OCI_BIN=$(if podman ps >/dev/null 2>&1; then echo podman; elif docker ps >/dev/null 2>&1; then echo docker; else echo "Neither podman nor docker is available." >&2; exit 1; fi)
    fi

    for node in $(./cluster/kubectl.sh get node --no-headers  -o custom-columns=":metadata.name"); do
      $OCI_BIN exec -t $node bash -c "echo 'fs.inotify.max_user_watches=1048576' >> /etc/sysctl.conf"
      $OCI_BIN exec -t $node bash -c "echo 'fs.inotify.max_user_instances=512' >> /etc/sysctl.conf"
      $OCI_BIN exec -i $node bash -c "sysctl -p /etc/sysctl.conf"
      if [[ "${node}" =~ worker ]]; then
          ./cluster/kubectl.sh label nodes $node node-role.kubernetes.io/worker="" --overwrite=true
      fi
    done
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
    export KUBECONFIG=${TMP_COMPONENT_PATH}/.output/kubeconfig
    export KIND_ARGS="-ic -i6 -mne -nse"
    make cluster-up

    trap teardown EXIT

    increase_ulimit
    cd ${TMP_PROJECT_PATH}
    export KUBEVIRT_PROVIDER=external
    export DEV_IMAGE_REGISTRY=localhost:5000
    ./cluster/cert-manager-install.sh
    deploy_cnao
    deploy_cnao_cr
    echo "Simulate network restrictions on CNAO namespace"
    ./hack/install-network-policy.sh
    ./hack/deploy-kubevirt.sh
    ./cluster/kubectl.sh -n kubevirt patch kubevirt kubevirt --type=json --patch '[{"op":"add","path":"/spec/configuration/developerConfiguration","value":{"featureGates":[]}},{"op":"add","path":"/spec/configuration/developerConfiguration/featureGates/-","value":"NetworkBindingPlugins"},{"op":"add","path":"/spec/configuration/developerConfiguration/featureGates/-","value":"DynamicPodInterfaceNaming"}]'
    ./cluster/kubectl.sh -n kubevirt patch kubevirt kubevirt --type=json --patch '[{"op":"add","path":"/spec/configuration/network","value":{"binding":{"l2bridge":{"domainAttachmentType":"managedTap","migration":{}}}}}]'
    ./cluster/kubectl.sh -n kubevirt patch kubevirt kubevirt --type=merge --patch '{"spec":{"configuration":{"virtualMachineOptions":{"disableSerialConsoleLog":{}}}}}'

    cd ${TMP_COMPONENT_PATH}
    echo "Run kubevirt-ipam-controller functional tests"
    make test-e2e
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
