#!/bin/bash -e

echo 'Cleaning up ...'

./cluster/kubectl.sh delete --ignore-not-found -f _out/namespace.yaml
./cluster/kubectl.sh delete --ignore-not-found -f _out/crds/network-addons-config.crd.yaml
./cluster/kubectl.sh delete --ignore-not-found -f _out/operator.yaml

sleep 2

echo 'Done'
