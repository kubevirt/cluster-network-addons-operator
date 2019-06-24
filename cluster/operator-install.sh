#!/bin/bash -e

./cluster/kubectl.sh create -f _out/cluster-network-addons/${VERSION}/operator.yaml
./cluster/kubectl.sh -n cluster-network-addons-operator wait deployment cluster-network-addons-operator --for condition=Available --timeout=600s
