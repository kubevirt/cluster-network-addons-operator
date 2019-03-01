# Cluster Network Addons Operator

Try it out:

```shell
./cluster/up.sh
./cluster/sync.sh
./cluster/kubectl.sh apply -f config-example.yaml
./cluster/kubectl.sh get pods --all-namespaces

./cluster/down.sh
```

`NetworkAddonsConfig` example:

```yaml
apiVersion: networkaddonsoperator.network.kubevirt.io/v1alpha1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  multus:
    delegates: |
      [{
        "type": "flannel",
        "name": "flannel.1",
        "delegate": {
          "isDefaultGateway": true
        }
      }]
```
