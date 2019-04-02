#!/bin/bash -e

echo 'Cleaning up ...'

./cluster/kubectl.sh delete --ignore-not-found -f _out/cluster-network-addons-operator_01_crd.yaml
./cluster/kubectl.sh delete --ignore-not-found -f _out/cluster-network-addons-operator_02_rbac.yaml
./cluster/kubectl.sh delete --ignore-not-found -f _out/cluster-network-addons-operator_00_namespace.yaml

sleep 2

echo 'Done'
