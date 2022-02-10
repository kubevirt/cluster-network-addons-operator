#!/usr/bin/env bash

set -xeuE

# This script should be able to execute multus
# functional tests against Kubernetes cluster with
# CNAO built with latest changes, on any
# environment with basic dependencies listed in
# check-patch.packages installed and docker running.
#
# Usage:
# yum -y install automation/check-patch.packages
# automation/check-patch.e2e-multus-functests.sh

teardown() {
    rm -rf "${TMP_COMPONENT_PATH}"
    cd ${TMP_PROJECT_PATH}
    make cluster-down
}

function __get-tools() {
    mkdir -p ${TMP_COMPONENT_PATH}/e2e/bin

    JQ_PATH=${TMP_COMPONENT_PATH}/e2e/bin/jq
    if ! [ -f ${JQ_PATH} ]; then
        echo "install jq"
        curl -Lo ${JQ_PATH} https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64
        chmod +x ${JQ_PATH}
    fi
}

function __prepare-test-environment() {
    __get-tools

    export CLUSTER_PATH=${TMP_PROJECT_PATH}/_kubevirtci/
    export KUBECTL=${TMP_PROJECT_PATH}/cluster/kubectl.sh

    echo "Deploy cni-plugins pod"
    CNI_PLUGIN_YAML="cni-install.yml"
    $KUBECTL create -f ${CNI_PLUGIN_YAML}
    $KUBECTL -n kube-system wait --for=condition=ready -l name=cni-plugins pod --timeout=300s

    TEST_YAML="*macvlan1.yml"
    # matching the test nodes to the CNAO cluster nodes
    sed -i 's/kind-worker$/node01/g' ${TEST_YAML}
    sed -i 's/kind-worker2$/node02/g' ${TEST_YAML}

    echo "Set eth1 NIC to up using NodeNetworkConfigurationPolicy"
    cat <<EOF | $KUBECTL apply -f -
apiVersion: nmstate.io/v1beta1
kind: NodeNetworkConfigurationPolicy
metadata:
  name: eth1-policy
spec:
  desiredState:
    interfaces:
    - name: eth1
      description: Configuring eth1 on all nodes
      type: ethernet
      state: up
      ipv4:
        dhcp: true
        enabled: true
EOF
}

main() {
    # Setup CNAO and artifacts temp directory
    source automation/check-patch.setup.sh
    cd ${TMP_PROJECT_PATH}

    # Spin-up ephemeral cluster with latest CNAO, 2 nodes and secondary NIC needed for test
    export KUBEVIRT_NUM_NODES=2
    export KUBEVIRT_NUM_SECONDARY_NICS=1
    # this script also exports KUBECONFIG, and fetch $COMPONENT repository
    COMPONENT="multus" source automation/components-functests.setup.sh

    trap teardown EXIT SIGINT

    # Run multus functional tests
    (
        cd ${TMP_COMPONENT_PATH}/e2e

        __prepare-test-environment

        echo "Run multus macvlan test"
        ./test-simple-macvlan1.sh
    )
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
