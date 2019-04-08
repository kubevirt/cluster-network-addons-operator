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
  sriov: {}
  kubeMacPool:
   startPoolRange: "02:00:00:00:00:00"
   endPoolRange: "FD:FF:FF:FF:FF:FF"
  imagePullPolicy: Always
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

## SR-IOV

The operator allows administrator to deploy SR-IOV
[device plugin](https://github.com/intel/sriov-network-device-plugin/) and
[CNI plugin](https://github.com/intel/sriov-cni/) simply by adding `sriov`
attribute to `NetworkAddonsConfig`.

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1alpha1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  sriov: {}
```

By default, device plugin is deployed with a configuration file with no root
devices configured, meaning that the plugin won't discover and advertise any
SR-IOV capable network devices.

If all nodes in the cluster are homogenous, meaning they have the same root
device IDs (or perhaps it's a single node deployment), then you can use
`SRIOV_ROOT_DEVICES` variable to specify appropriate IDs. If they are not, you
can deploy with default configuration file and then modify
`/etc/pcidp/config.json` on each node to list corresponding root device IDs.
You may need to restart SR-IOV device plugin pods to catch up configuration
file changes.

Additionally, container images used to deliver these plugins can be set using
`SRIOV_DP_IMAGE` and `SRIOV_CNI_IMAGE` environment variables in operator
deployment manifest.

**Note:** OpenShift 4 is shipped with [Cluster Network
Operator](https://github.com/openshift/cluster-network-operator). OpenShift
operator already supports SR-IOV deployment. But it uses older versions of
components that are not compatible with KubeVirt SR-IOV feature. Therefore, if
SR-IOV is requested in OpenShift cluster network operator, KubeVirt addons
operator will return an error.

**Note:** To use SR-IOV for KubeVirt, one should also create a corresponding
network attachment definition resource. For example:

```yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: sriov-net1
  annotations:
    k8s.v1.cni.cncf.io/resourceName: intel.com/sriov
spec:
  config: '{
  "type": "sriov",
  "name": "sriov-network",
  "ipam": {
    "type": "host-local",
    "subnet": "10.56.217.0/24",
    "routes": [{
      "dst": "0.0.0.0/0"
    }],
    "gateway": "10.56.217.1"
  }
}'
```

## Kubemacpool
The operator allows administrator to deploy the [Kubemacpool](https://github.com/K8sNetworkPlumbingWG/kubemacpool)
This project allow to allocate mac addresses from a pool to secondary interfaces using
[Network Plumbing Working Group de-facto standard](https://github.com/K8sNetworkPlumbingWG/multi-net-spec).

Administrator need to specify a requested range

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1alpha1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  kubeMacPool:
   startPoolRange: "02:00:00:00:00:00"
   endPoolRange: "FD:FF:FF:FF:FF:FF"
```

## Image Pull Policy

Administrator can specify [image pull policy](https://kubernetes.io/docs/concepts/containers/images/)
for deployed components. Default is `IfNotPresent`.

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1alpha1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  imagePullPolicy: Always
```

# Deployment

First install the operator itself:

```shell
kubectl apply -f https://raw.githubusercontent.com/kubevirt/cluster-network-addons-operator/master/deploy/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/kubevirt/cluster-network-addons-operator/master/deploy/crds/network-addons-config.crd.yaml
kubectl apply -f https://raw.githubusercontent.com/kubevirt/cluster-network-addons-operator/master/deploy/operator.yaml
```

Then you need to create a configuration for the operator [example
CR](deploy/crds/network-addons-config-example.cr.yaml):

```shell
kubectl apply -f https://raw.githubusercontent.com/kubevirt/cluster-network-addons-operator/master/deploy/crds/network-addons-config-example.cr.yaml
```

For more information about the configuration format check [configuring section](#configuration).

# Development

```shell
# validate imports
make vet

# validate formatting
make fmt

# generate manifests
make manifests

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

## Deploy Using OLM

For more information on the [Operator Lifecycle
Manager](https://github.com/operator-framework/operator-lifecycle-manager) and
Cluster Service Versions checkout out ["Building a Cluster Service
Version"](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/Documentation/design/building-your-csv.md).

**NOTE**

Replace `<docker_org>` with your Docker organization.

1) Build and push an operator-registry image.

```shell
cd deploy
export DOCKER_ORG=<docker_org>
docker build --no-cache -t docker.io/$DOCKER_ORG/cna-registry:example -f Dockerfile .
docker push docker.io/$DOCKER_ORG/cna-registry:example
```

2) Create the cluster-network-addons-operator Namespace and OperatorGroup.
[This](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/Documentation/design/building-your-csv.md#operator-metadata)
portion of OLM's CSV writing documentation goes into more detail about
OperatorGroups.

```shell
kubectl create ns cluster-network-addons-operator

cat <<EOF | kubectl create -f -
apiVersion: operators.coreos.com/v1alpha2
kind: OperatorGroup
metadata:
  name: cluster-network-addons-operator
  namespace: cluster-network-addons-operator
EOF
```

3) Using the `cna-registry` container image built in step 1,
create a CatalogSource. This object tells OLM about the operator.

```bash
cat <<EOF | kubectl create -f -
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: cluster-network-addons
  namespace: openshift-operator-lifecycle-manager
spec:
  sourceType: grpc
  image: docker.io/$DOCKER_ORG/cna-registry:example
  displayName: Cluster Network Addons
  publisher: Red Hat
EOF
```

4) Subscribe to the cluster-network-addons-operator.

```shell
cat <<EOF | kubectl create -f -
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: cluster-network-addons-operator-subscription
  namespace: cluster-network-addons-operator
spec:
  channel: alpha
  name: cluster-network-addons
  source: cluster-network-addons
  sourceNamespace: openshift-operator-lifecycle-manager
  startingCSV: cluster-network-addons-operator.v0.0.0
EOF
```

5) With the cluster-network-addons-operator ready and available, create a NetworkAddonsConfig.

```shell
cat <<EOF | kubectl create -f -
apiVersion: networkaddonsoperator.network.kubevirt.io/v1alpha1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  imagePullPolicy: Always
  kubeMacPool:
    endPoolRange: FD:FF:FF:FF:FF:FF
    startPoolRange: "02:00:00:00:00:00"
  linuxBridge: {}
  multus: {}
  sriov: {}
```
