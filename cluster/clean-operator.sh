#!/bin/bash -e

echo 'Cleaning up ...'

./cluster/kubectl.sh delete --ignore-not-found -f _out_operator/cluster-network-addons/${VERSION}/namespace.yaml
./cluster/kubectl.sh delete --ignore-not-found -f _out_operator/cluster-network-addons/${VERSION}/network-addons-config.crd.yaml
./cluster/kubectl.sh delete --ignore-not-found -f _out_operator/cluster-network-addons/${VERSION}/operator.yaml

sleep 2

echo 'Done'
