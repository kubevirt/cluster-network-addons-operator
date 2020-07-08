#!/usr/bin/env bash

set -xeo pipefail

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh

echo 'Bumping ovs-cni'
OVS_URL=$(yaml-utils::get_component_url ovs-cni)
OVS_COMMIT=$(yaml-utils::get_component_commit ovs-cni)
OVS_REPO=$(yaml-utils::get_component_repo ${OVS_URL})

TEMP_DIR=$(git-utils::create_temp_path ovs-cni)
trap "rm -rf ${TEMP_DIR}" EXIT
OVS_PATH=${TEMP_DIR}/${OVS_REPO}

echo 'Fetch ovs-cni sources'
git-utils::fetch_component ${OVS_PATH} ${OVS_URL} ${OVS_COMMIT}

OVS_TAG=$(git-utils::get_component_tag ${OVS_PATH})

echo 'Get ovs-cni-plugin image name and update it under CNAO'
OVS_PLUGIN_IMAGE=quay.io/kubevirt/ovs-cni-plugin
OVS_PLUGIN_IMAGE_TAGGED=${OVS_PLUGIN_IMAGE}:${OVS_TAG}
sed -i "s#\"${OVS_PLUGIN_IMAGE}:.*\"#\"${OVS_PLUGIN_IMAGE_TAGGED}\"#" \
    pkg/components/components.go \
    test/releases/${CNAO_VERSION}.go

echo 'Get ovs-cni-marker image name and update it under CNAO'
OVS_MARKER_IMAGE=quay.io/kubevirt/ovs-cni-marker
OVS_MARKER_IMAGE_TAGGED=${OVS_MARKER_IMAGE}:${OVS_TAG}
sed -i "s#\"${OVS_MARKER_IMAGE}:.*\"#\"${OVS_MARKER_IMAGE_TAGGED}\"#" \
    pkg/components/components.go \
    test/releases/${CNAO_VERSION}.go
