#!/bin/bash -e

registry_port=$(./cluster/cli.sh ports registry | tr -d '\r')
registry=localhost:$registry_port

./cluster/clean.sh

make docker-build docker-push OPERATOR_IMAGE_REGISTRY=$registry

for i in $(seq 1 ${CLUSTER_NUM_NODES}); do
    ./cluster/cli.sh ssh "node$(printf "%02d" ${i})" 'sudo docker pull registry:5000/cluster-network-addons-operator:latest'
    # Temporary until image is updated with provisioner that sets this field
    # This field is required by buildah tool
    ./cluster/cli.sh ssh "node$(printf "%02d" ${i})" 'sudo sysctl -w user.max_user_namespaces=1024'
done

make generate-manifests OPERATOR_IMAGE_REGISTRY=registry:5000 MANIFESTS_OUTPUT_DIR=_out/

./cluster/kubectl.sh create -f _out/cluster-network-addons-operator_00_namespace.yaml
./cluster/kubectl.sh create -f _out/cluster-network-addons-operator_01_crd.yaml
./cluster/kubectl.sh create -f _out/cluster-network-addons-operator_02_rbac.yaml
./cluster/kubectl.sh create -f _out/cluster-network-addons-operator_03_deployment.yaml
