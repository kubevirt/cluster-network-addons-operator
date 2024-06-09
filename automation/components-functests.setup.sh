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

# Spin up Kubernetes cluster
make cluster-down cluster-up

# Export .kubeconfig full path, so it will be possible
# to use 'kubectl' directly from the component directory path
export KUBECONFIG=$(cluster::kubeconfig)

# Deploy CNAO latest changes
make cluster-operator-push
make cluster-operator-install

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

# Deploy all network addons components with CNAO
    cat <<EOF | cluster/kubectl.sh apply -f -
apiVersion: networkaddonsoperator.network.kubevirt.io/v1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  multus: {}
  multusDynamicNetworks: {}
  linuxBridge: {}
  kubeMacPool:
   rangeStart: "02:00:00:00:00:00"
   rangeEnd: "02:00:00:00:00:0F"
  ovs: {}
  macvtap: {}
  kubeSecondaryDNS: {}
  kubevirtIpamController: {}
  imagePullPolicy: Always
EOF

if [[ ! $(cluster/kubectl.sh wait networkaddonsconfig cluster --for condition=Available --timeout=13m) ]]; then
	echo "Failed to wait for CNAO CR to be ready"
	cluster/kubectl.sh get networkaddonsconfig -o custom-columns="":.status.conditions[*].message
	exit 1
fi

# Clone component repository
component_url=$(yaml-utils::get_component_url ${COMPONENT})
component_commit=$(yaml-utils::get_component_commit ${COMPONENT})
component_repo=$(yaml-utils::get_component_repo ${component_url})

component_temp_dir=$(git-utils::create_temp_path ${COMPONENT})
component_path=${component_temp_dir}/${component_repo}

git-utils::fetch_component ${component_path} ${component_url} ${component_commit}

export TMP_COMPONENT_PATH=${component_path}
