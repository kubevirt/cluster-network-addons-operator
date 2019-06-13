#!/bin/bash -e

registry_port=$(./cluster/cli.sh ports registry | tr -d '\r')
registry=localhost:$registry_port

# Cleanup previously generated manifests
rm -rf _out_operator/
# Copy release manifests as a base for generated ones, this should make it possible to upgrade
cp -r manifests/ _out_operator/
IMAGE_REGISTRY=registry:5000 DEPLOY_DIR=_out_operator make gen-manifests

make cluster-clean

IMAGE_REGISTRY=$registry make docker-build-operator docker-push-operator

for i in $(seq 1 ${CLUSTER_NUM_NODES}); do
    ./cluster/cli.sh ssh "node$(printf "%02d" ${i})" 'sudo docker pull registry:5000/cluster-network-addons-operator'
    # Temporary until image is updated with provisioner that sets this field
    # This field is required by buildah tool
    ./cluster/cli.sh ssh "node$(printf "%02d" ${i})" 'sudo sysctl -w user.max_user_namespaces=1024'
done

./cluster/kubectl.sh create -f _out_operator/cluster-network-addons/${VERSION}/namespace.yaml
./cluster/kubectl.sh create -f _out_operator/cluster-network-addons/${VERSION}/network-addons-config.crd.yaml
./cluster/kubectl.sh create -f _out_operator/cluster-network-addons/${VERSION}/operator.yaml
./cluster/kubectl.sh -n cluster-network-addons-operator wait deployment cluster-network-addons-operator --for condition=Available --timeout=600s
