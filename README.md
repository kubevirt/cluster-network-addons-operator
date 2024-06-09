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
  multusDynamicNetworks: {}
  linuxBridge: {}
  kubeMacPool: {}
  ovs: {}
  macvtap: {}
  kubeSecondaryDNS: {}
  kubevirtIpamController: {}
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

## Multus Dynamic Networks Controller

[This controller](https://github.com/k8snetworkplumbingwg/multus-dynamic-networks-controller)
allows hot-plug and hot-unplug of additional Pod intefaces. It is done using
`multusDynamicNetworks` attribute.

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  multus: {}
  multusDynamicNetworks: {}
```

Additionally, container image used to deliver this plugin can be set using
`MULTUS_DYNAMIC_NETWORKS_CONTROLLER_IMAGE` environment variable in operator
deployment manifest.

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

## NMState

**Note:** The cluster-network-addons-operator is no longer installing
kubernetes-nmstate, refer to its own operator [release notes](https://github.com/nmstate/kubernetes-nmstate/releases) to install it.

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

A simple example on how to do so, the user must deploy a `ConfigMap`, such as in [this example](https://github.com/kubevirt/macvtap-cni/blob/main/examples/macvtap-deviceplugin-config-explicit.yaml).

Currently, this configuration is not dynamic.

## KubeSecondaryDNS

[This controller](https://github.com/kubevirt/kubesecondarydns)
allows to support FQDN for VMI's secondary networks.

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  kubeSecondaryDNS:
    DOMAIN: ""
    NAME_SERVER_IP: ""
```

Additionally, container image used to deliver this plugin can be set using
`KUBE_SECONDARY_DNS_IMAGE` environment variable in operator
deployment manifest.

## kubevirtIpamController

[This controller](https://github.com/maiqueb/kubevirt-ipam-claims)
allows to support IPAM for secondary networks.

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  multus: {}
  kubevirtIpamController: {}
```

Additionally, container image used to deliver this plugin can be set using
`KUBEVIRT_IPAM_CONTROLLER_IMAGE` environment variable in operator
deployment manifest.

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

Administrator can specify [webhook self signed certificates configuration](https://pkg.go.dev/github.com/qinqon/kube-admission-webhook@v0.13.0/pkg/certificate?tab=doc#Options)
for deployed components. Default is `caRotateInterval: 168h`, `caOverlapInterval: 24h`, `certRotateInterval: 24h`, `certOverlapInterval: 8h`

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  selfSignConfiguration:
    caRotateInterval: 168h
    caOverlapInterval: 24h
    certRotateInterval: 24h
    certOverlapInterval: 8h
```
The selfSignConfiguration parameters has to be all or none set, setting some of
them fails at validation. They have to conform to golang time.Duration
string format. Additionally the following checks are done at validation:
- caRotateInterval >= caOverlapInterval && caRotateInterval >= certRotateInterval && certRotateInterval >= certOverlapInterval

This parameters are consumed by Kubemacpool component.

## Placement Configuration

CNAO deploys two component categories: infra and workload. Workload components manage node configuration
and need to be scheduled on the nodes where actual user workload is scheduled.
Infra components provide cluster-wide service and do not need to be running on the same nodes as user workload.

Administrator can specify placement preferences for deployed infra and workload components
by defining [affinity, nodeSelector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/)
and [tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/).

By default, infra components are scheduled on control-plane nodes and workload components are scheduled on all nodes.
To adjust this behaviour, provide custom `placementConfiguration` to the `NetworkAddonsConfig`.

In the following example, `nodeAffinity` is used to schedule infra components to control-plane nodes and `nodeSelector`
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
              - key: node-role.kubernetes.io/control-plane
                operator: Exists
    workloads:
      nodeSelector:
        node-role.kubernetes.io/worker: ""
```

# Deployment

First install the operator itself:

```shell
kubectl apply -f https://github.com/kubevirt/cluster-network-addons-operator/releases/download/v0.93.0/namespace.yaml
kubectl apply -f https://github.com/kubevirt/cluster-network-addons-operator/releases/download/v0.93.0/network-addons-config.crd.yaml
kubectl apply -f https://github.com/kubevirt/cluster-network-addons-operator/releases/download/v0.93.0/operator.yaml
```

Then you need to create a configuration for the operator [example
CR](manifests/cluster-network-addons/0.4.0/network-addons-config-example.cr.yaml):

```shell
kubectl apply -f https://github.com/kubevirt/cluster-network-addons-operator/releases/download/v0.93.0/network-addons-config-example.cr.yaml
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

Starting with version `0.76.0`, this operator supports upgrades to any newer
version. If you wish to upgrade, remove old operator (`operator.yaml`) and
install new, operands will remain available during the operator's downtime.

# Development

Make sure you have either Docker >= 17.05 / podman >= 3.1 installed.

```shell
# run code validation and unit tests
make check

# perform auto-formatting on the source code (if not done by your IDE)
make fmt

# generate source code for API
make gen-k8s

# build images (uses multi-stage builds and therefore requires Docker >= 17.05 / podman >= 3.1)
make docker-build

# or build only a specific image
make docker-build-operator
make docker-build-registry

# bring up a local cluster with Kubernetes
make cluster-up

# bridge up a local cluster with kubernetes 1.25
export KUBEVIRT_PROVIDER=k8s-1.25
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
