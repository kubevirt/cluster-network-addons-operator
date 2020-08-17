# Bumper Script

The Bumper script goes over all of CNAO's components, using the components.yaml config, finds new releases and bumps them in separate PRs. The script can be run locally or via an automation such as GitHub Actions.

## How to run the script

```
go build $GOPATH/src/github.com/kubevirt/cluster-network-addons-operator/tools/bumper
./bumper -config-path=<path-to-components.yaml> -token=<git-token>
```
