#!/usr/bin/env bash
set -xe

make cluster-down
make cluster-up
make cluster-sync

NEW_VERSION=99.0.0
OLD_VERSION=0.16.0

echo "1. deploy version $OLD_VERSION"

kubectl apply -f _out/cluster-network-addons/${OLD_VERSION}/namespace.yaml
kubectl apply -f _out/cluster-network-addons/${OLD_VERSION}/network-addons-config.crd.yaml
kubectl apply -f _out/cluster-network-addons/${OLD_VERSION}/operator.yaml

kubectl -n cluster-network-addons wait deployment cluster-network-addons-operator --for condition=Available --timeout=600s

kubectl apply -f _out/cluster-network-addons/${OLD_VERSION}/network-addons-config-example.cr.yaml

sleep 120

echo "2. deploy version $NEW_VERSION"
kubectl apply -f _out/cluster-network-addons/${NEW_VERSION}/network-addons-config.crd.yaml
kubectl apply -f _out/cluster-network-addons/${NEW_VERSION}/operator.yaml

kubectl -n cluster-network-addons wait deployment cluster-network-addons-operator --for condition=Available --timeout=600s

kubectl apply -f _out/cluster-network-addons/${NEW_VERSION}/network-addons-config-example.cr.yaml

sleep 120

echo "3. remove version $NEW_VERSION"
kubectl delete -f _out/cluster-network-addons/${NEW_VERSION}/network-addons-config.crd.yaml
kubectl delete -f _out/cluster-network-addons/${NEW_VERSION}/operator.yaml

sleep 120

echo "4. re-deploy version $OLD_VERSION"
kubectl apply -f _out/cluster-network-addons/${OLD_VERSION}/network-addons-config.crd.yaml
kubectl apply -f _out/cluster-network-addons/${OLD_VERSION}/operator.yaml

kubectl -n cluster-network-addons wait deployment cluster-network-addons-operator --for condition=Available --timeout=600s
kubectl apply -f _out/cluster-network-addons/${OLD_VERSION}/network-addons-config-example.cr.yaml
