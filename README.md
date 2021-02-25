# Cluster Network Addons Operator

This operator can be used to deploy additional networking components on top of
Kubernetes cluster.

# Configuration

Configuration of desired network addons is done using `NetworkAddonsConfig` object:

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  multus: {}
  linuxBridge: {}
  kubeMacPool: {}
  nmstate: {}
  ovs: {}
  macvtap: {}
  imagePullPolicy: Always
```

## Multus

The operator allows administrator to deploy multi-network
[Multus plugin](https://github.com/intel/multus-cni). It is done using `multus`
attribute.

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  multus: {}
```

Additionally, container image used to deliver this plugin can be set using
`MULTUS_IMAGE` environment variable in operator deployment manifest.

## Linux Bridge

The operator allows administrator to deploy [Linux Bridge CNI plugin](https://github.com/containernetworking/plugins/tree/master/plugins/main/bridge)
simply by adding `linuxBridge` attribute to `NetworkAddonsConfig`.

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1
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
apiVersion: networkaddonsoperator.network.kubevirt.io/v1
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
apiVersion: networkaddonsoperator.network.kubevirt.io/v1
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
apiVersion: networkaddonsoperator.network.kubevirt.io/v1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  ovs: {}
```

## Macvtap

**Note:** This feature is **experimental**. Macvtap-cni is unstable and its API
may change.

The operator allows the administrator to deploy the
[macvtap CNI plugin](https://github.com/kubevirt/macvtap-cni/), simply by
adding `macvtap` attribute to `NetworkAddonsConfig`.

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  macvtap: {}
```

Macvtap-cni must be explicitly configured by the administrator, indicating the
interfaces on top of which logical networks can be created.

A simple example on how to do so, the user must deploy a `ConfigMap`, such as in [this example](https://github.com/kubevirt/macvtap-cni/blob/master/examples/macvtap-deviceplugin-config.yaml).

Currently, this configuration is not dynamic.

## Image Pull Policy

Administrator can specify [image pull policy](https://kubernetes.io/docs/concepts/containers/images/)
for deployed components. Default is `IfNotPresent`.

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  imagePullPolicy: Always
```

## Self Signed Certificates Configuration

Administrator can specify [webhook self signed certificates configuration](https://pkg.go.dev/github.com/qinqon/kube-admission-webhook@v0.12.0/pkg/certificate?tab=doc#Options)
for deployed components. Default is `caRotateInterval: 168h`, `caOverlapInterval: 7h`, `certRotateInterval: 7h`

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  selfSignConfiguration:
    caRotateInterval: 168h
    caOverlapInterval: 7h
    certRotateInterval: 7h
```
The selfSignConfiguration parameters has to be all or none set, setting some of
them fails at validation, also they have to conform to golang time.Duration
string format also the following checks are done at validation: caRotateInterval => caOverlapInterval && caRotateInterval => certRotateInterval

This parameters are consumed by kubemacpool and kubernetes-nmstate components.

## Placement Configuration

Administrator can specify placement preferences for deployed infra and workload components
by defining [affinity, nodeSelector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/)
and [tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/).

By default, infra components are scheduled on master nodes and workload components are scheduled on all nodes.
To adjust this behaviour, provide custom `placementConfiguration` to the `NetworkAddonsConfig`.

In the following example, `nodeAffinity` is used to schedule infra components to master nodes and `nodeSelector`
to schedule workloads on worker nodes.
Note that worker nodes need to be labeled with `node-role.kubernetes.io/worker` label.

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  placementConfiguration:
    infra:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: node-role.kubernetes.io/master
                operator: Exists
    workloads:
      nodeSelector:
        node-role.kubernetes.io/worker: ""
```

# Deployment

First install the operator itself:

```shell
kubectl apply -f https://github.com/kubevirt/cluster-network-addons-operator/releases/download/v0.44.4/namespace.yaml
kubectl apply -f https://github.com/kubevirt/cluster-network-addons-operator/releases/download/v0.44.4/network-addons-config.crd.yaml
kubectl apply -f https://github.com/kubevirt/cluster-network-addons-operator/releases/download/v0.44.4/operator.yaml
```

Then you need to create a configuration for the operator [example
CR](manifests/cluster-network-addons/0.4.0/network-addons-config-example.cr.yaml):

```shell
kubectl apply -f https://github.com/kubevirt/cluster-network-addons-operator/releases/download/v0.44.4/network-addons-config-example.cr.yaml
```

Finally you can wait for the operator to finish deployment:

```shell
kubectl wait networkaddonsconfig cluster --for condition=Available
```

In case something failed, you can find the error in the NetworkAddonsConfig Status field:

```shell
kubectl get networkaddonsconfig cluster -o yaml
```
You can follow the deployment state through events produced in the default namespace:

```shell
kubectl get events
```
Events will be produced whenever the deployment is applied, configured or failed. The expected events are:

|Event type   | Reason                                                            |
|-------------|-------------------------------------------------------------------|
|Progressing  | When operator had started deploying the components                |
|Failed       | When one or more components failed to deploy                      |
|Available    | When all components finished to deploy                            |
|Modified     | When the configuration was modified or applied for the first time |



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

# bridge up a local cluster with kubernetes 1.19
export KUBEVIRT_PROVIDER=k8s-1.19
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

For developing at an external cluster:

```bash
export KUBEVIRT_PROVIDER=external

export KUBECONFIG=[path to external's cluster kubeconfig]

# This is the registry used to push and pull the dev image
# it has to be accessible by the external cluster
export DEV_IMAGE_REGISTRY=quay.io/$USER

# Then is possible to follow normal dev flow

make cluster-operator-push
make cluster-operator-install
make test/e2e/workflow
make test/e2e/lifecycle
```

# Releasing

1. Checkout a public branch
2. Call `make prepare-patch|minor|major` and prepare release notes
3. Open a new PR
4. Once the PR is merged, create a new release in GitHub and attach new manifests
