#! /bin/bash
set -exu

# Create an ephemeral cluster with latest changes
# CNAO and clones the desired component repository
#
# This script exports:
# - KUBECONFIG
#   kubeconfig full path so you can use`kubectl` binary or
#   CNAO's `./cluster/kubectl.sh` script directly.
#
# - TMP_COMPONENT_PATH
#   The desired component temp directory
#
# Example:
# COMPONENT="kubemacpool" source automation/e2e-functest.setup.sh
# cd ${TMP_COMPONENT_PATH}
# KUBECONFIG=$KUBECONFIG make functests

source hack/components/git-utils.sh
source hack/components/yaml-utils.sh
source cluster/cluster.sh

USE_KUBEVIRTCI=${USE_KUBEVIRTCI:-"true"}

# Export .kubeconfig full path, so it will be possible
# to use 'kubectl' directly from the component directory path
export KUBECONFIG=${KUBECONFIG:-$(cluster::kubeconfig)}

function deploy_cluster {
  # Spin up Kubernetes cluster
  export KUBEVIRT_MEMORY_SIZE=10240M
  make cluster-down cluster-up
}

function deploy_cnao {
  # Deploy CNAO latest changes
  make cluster-operator-push
  make cluster-operator-install
}

function patch_restricted_namespace {
  # Test kubemacpool with restricted
  if [ "$COMPONENT" == "kubemacpool" ]; then
    cluster/kubectl.sh apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: cluster-network-addons
  labels:
    pod-security.kubernetes.io/enforce: restricted
EOF
  fi
}

function deploy_cnao_cr {
  # Deploy all network addons components with CNAO

  cat <<EOF > cr.yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  kubeMacPool:
   rangeStart: "02:00:00:00:00:00"
   rangeEnd: "02:00:00:00:00:0F"
  kubevirtIpamController: {}
  imagePullPolicy: Always
EOF

  if [[ $USE_KUBEVIRTCI == true ]]; then
    echo "  linuxBridge: {}" >> cr.yaml
    echo "  multus: {}" >> cr.yaml
    echo "  multusDynamicNetworks: {}" >> cr.yaml
    echo "  ovs: {}" >> cr.yaml
    echo "  macvtap: {}" >> cr.yaml
    echo "  kubeSecondaryDNS: {}" >> cr.yaml
  fi

  cluster/kubectl.sh apply -f cr.yaml

  if [[ ! $(cluster/kubectl.sh wait networkaddonsconfig cluster --for condition=Available --timeout=13m) ]]; then
    echo "Failed to wait for CNAO CR to be ready"
    cluster/kubectl.sh get networkaddonsconfig -o custom-columns="":.status.conditions[*].message
    exit 1
  fi
}

# Clone component repository
component_url=$(yaml-utils::get_component_url ${COMPONENT})
component_commit=$(yaml-utils::get_component_commit ${COMPONENT})
component_repo=$(yaml-utils::get_component_repo ${component_url})

component_temp_dir=$(git-utils::create_temp_path ${COMPONENT})
component_path=${component_temp_dir}/${component_repo}

git-utils::fetch_component ${component_path} ${component_url} ${component_commit}

export TMP_COMPONENT_PATH=${component_path}

if [[ $USE_KUBEVIRTCI == true ]]; then
  deploy_cluster
  deploy_cnao
  patch_restricted_namespace
  deploy_cnao_cr
fi
