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
  kubeMacPool: {}
  nmstate: {}
  ovs: {}
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

The bridge marker image used to deliver a bridge marker detecting the availability
of linux bridges on nodes can be set using the `LINUX_BRIDGE_MARKER_IMAGE` environment
variable in operator deployment manifest.

### Configure bridge on node

Following snippets can be used to configure linux bridge on your node.

```shell
# create the bridge using NetworkManager
nmcli con add type bridge ifname br10

# allow traffic to go through the bridge between pods
iptables -I FORWARD 1 -i br10 -j ACCEPT
```

## Kubemacpool
The operator allows administrator to deploy the [Kubemacpool](https://github.com/K8sNetworkPlumbingWG/kubemacpool).
This project allow to allocate mac addresses from a pool to secondary interfaces using
[Network Plumbing Working Group de-facto standard](https://github.com/K8sNetworkPlumbingWG/multi-net-spec).

**Note:** Administrator can specify a requested range, if the range is not
requested a random range will be provided. This random range spans from
`02:XX:XX:00:00:00` to `02:XX:XX:FF:FF:FF`, where `02` makes the address local
unicast and `XX:XX` is a random prefix.

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1alpha1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  kubeMacPool:
   rangeStart: "02:00:00:00:00:00"
   rangeEnd: "FD:FF:FF:FF:FF:FF"
```

## NMState

**Note:** This feature is **experimental**. NMState is unstable and its API
may change.

The operator allows the administrator to deploy the [NMState State
Controller](https://github.com/nmstate/nmstate) as a daemonset across all of
one's nodes. This project manages host networking settings in a declarative
manner. The networking state is described by a pre-defined schema. Reporting of
current state and changes to it (desired state) both conform to the schema.
NMState is aimed to satisfy enterprise needs to manage host networking through a
northbound declarative API and multi provider support on the southbound.
NetworkManager acts as the main (and currently the only) provider supported.

This component can be enabled by adding `nmstate` section to the
`NetworkAddonsConfig`.

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1alpha1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  nmstate: {}
```

It communicate with a NetworkManager instance running on the node using D-Bus.
Make sure that NetworkManager is installed and running on each node.

```shell
yum install NetworkManager
systemctl start NetworkManager
```

## Open vSwitch

The operator allows administrator to deploy [OVS CNI plugin](https://github.com/kubevirt/ovs-cni/)
simply by adding `ovs` attribute to `NetworkAddonsConfig`. Please note that
in order to use this plugin, openvswitch have to be up and running at nodes.

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1alpha1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  ovs: {}
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
kubectl apply -f https://raw.githubusercontent.com/kubevirt/cluster-network-addons-operator/master/manifests/cluster-network-addons/0.27.0/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/kubevirt/cluster-network-addons-operator/master/manifests/cluster-network-addons/0.27.0/network-addons-config.crd.yaml
kubectl apply -f https://raw.githubusercontent.com/kubevirt/cluster-network-addons-operator/master/manifests/cluster-network-addons/0.27.0/operator.yaml
```

Then you need to create a configuration for the operator [example
CR](manifests/cluster-network-addons/0.4.0/network-addons-config-example.cr.yaml):

```shell
kubectl apply -f https://raw.githubusercontent.com/kubevirt/cluster-network-addons-operator/master/manifests/cluster-network-addons/0.27.0/network-addons-config-example.cr.yaml
```

Finally you can wait for the operator to finish deployment:

```shell
kubectl wait networkaddonsconfig cluster --for condition=Available
```

In case something failed, you can find the error in the NetworkAddonsConfig Status field:

```shell
kubectl get networkaddonsconfig cluster -o yaml
```

For more information about the configuration format check [configuring section](#configuration).

# Upgrades

Starting with version `0.16.0`, this operator supports upgrades to any newer
version. If you wish to upgrade, remove old operator (`operator.yaml`) and
install new, operands will remain available during the operator's downtime.

# Development

Make sure you have Docker >= 17.05 installed.

```shell
# run code validation and unit tests
make check

# perform auto-formatting on the source code (if not done by your IDE)
make fmt

# generate source code for API
make gen-k8s

# build images (uses multi-stage builds and therefore requires Docker >= 17.05)
make docker-build

# or build only a specific image
make docker-build-operator
make docker-build-registry

# bring up a local cluster with Kubernetes
make cluster-up

# bridge up a local cluster with OpenShift 3
export CLUSTER_PROVIDER='os-3.11.0'
make cluster-up

# build images and push them to the local cluster
make cluster-operator-push

# install operator on the local cluster
make cluster-operator-install

# run workflow e2e tests on the cluster, requires cluster with installed operator,
# workflow covers deployment of operands
make test/e2e/workflow

# run lifecycle e2e tests on the cluster, requires cluster without operator installed,
# lifecycle covers deployment of operator itself and its upgrades
make test/e2e/lifecycle

# access kubernetes API on the cluster
./cluster/kubectl.sh get nodes

# ssh into the cluster's node
./cluster/cli.sh ssh node01

# clean up all resources created by the operator from the cluster
make cluster-clean

# delete the cluster
make cluster-down
```

# Releasing

Steps to create a new release:

1. Test operator, make sure it deploys all components, exposes failures in NetworkAddonsConfig.Status field as well as progressing status of components and "Ready".
2. Open a new PR with two commits. The first of them adding new released manifests, the second bumping versions in `Makefile`. To make this easier, use `./hack/release.sh <previous version> <released version> <future version> <origin remote> <fork remote>`
3. Once the PR is merged, tag its **first** commit with proper version name `x.y.z`. This can be done through [GitHub UI](https://github.com/kubevirt/cluster-network-addons-operator/releases/new).
