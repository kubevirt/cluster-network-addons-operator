# Cluster Network Addons Operator

This operator can be used to deploy additional networking components on top of
Kubernetes/OpenShift cluster.

On OpenShift 4, it is preferred to use the native [OpenShift ClusterNetworkOperator](https://github.com/openshift/cluster-network-operator).
However, not all features might be supported there.

# Configuration

Configuration of desired network addons is done using `NetworkAddonsConfig` object:

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1alpha1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  multus: {}
  linuxBridge: {}
```

## Multus

The operator allows administrator to deploy multi-network
[Multus plugin](https://github.com/intel/multus-cni). It is done using `multus`
attribute.

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1alpha1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  multus: {}
```

Additionally, container image used to deliver this plugin can be set using
`MULTUS_IMAGE` environment variable in operator deployment manifest.

**Note:** OpenShift 4 is shipped with [Cluster Network Operator](https://github.com/openshift/cluster-network-operator). OpenShift operator already supports Multus deployment. Therefore, if Multus is requested in our operator using `multus` attribute, we just make sure that is is not disabled in the OpenShift one.

## Linux Bridge

The operator allows administrator to deploy [Linux Bridge CNI plugin](https://github.com/containernetworking/plugins/tree/master/plugins/main/bridge)
simply by adding `linuxBridge` attribute to `NetworkAddonsConfig`.

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1alpha1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  linuxBridge: {}
```

Additionally, container image used to deliver this plugin can be set using
`LINUX_BRIDGE_IMAGE` environment variable in operator deployment manifest.

# Deployment

First install the operator itself:

```shell
kubectl apply -f https://raw.githubusercontent.com/kubevirt/cluster-network-addons-operator/master/deploy/cluster-network-addons-operator_00_namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/kubevirt/cluster-network-addons-operator/master/deploy/cluster-network-addons-operator_01_crd.yaml
kubectl apply -f https://raw.githubusercontent.com/kubevirt/cluster-network-addons-operator/master/deploy/cluster-network-addons-operator_02_rbac.yaml
kubectl apply -f https://raw.githubusercontent.com/kubevirt/cluster-network-addons-operator/master/deploy/cluster-network-addons-operator_03_deployment.yaml
```

Then you need to create a configuration for the operator:

```yaml
cat <<EOF | kubectl create -f -
apiVersion: networkaddonsoperator.network.kubevirt.io/v1alpha1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  multus: {}
  linuxBridge: {}
EOF
```

For more information about the configuration format check [configuring section](#configuration).

# Development

```shell
# validate imports
make vet

# validate formatting
make fmt

# generate sources (requires operator-sdk installed on your host)
operator-sdk generate k8s

# build image (uses multi-stage builds and therefore requires Docker >= 17.05)
make docker-build

# bring up a local cluster with Kubernetes
make cluster-up

# bridge up a local cluster with OpenShift 3
export CLUSTER_PROVIDER='os-3.11.0'
make cluster-up

# deploy operator from sources on the cluster
make cluster-sync

# access kubernetes API on the cluster
./cluster/kubectl.sh get nodes

# ssh into the cluster's node
./cluster/cli.sh ssh node01

# clean up all resources created by the operator from the cluster
make cluster-clean

# delete the cluster
make cluster-down
```
