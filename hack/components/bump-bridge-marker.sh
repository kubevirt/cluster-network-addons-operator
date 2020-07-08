#!/usr/bin/env bash

set -xeo pipefail

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh

echo 'Bumping bridge-marker'
BRIDGE_MARKER_URL=$(yaml-utils::get_component_url bridge-marker)
BRIDGE_MARKER_COMMIT=$(yaml-utils::get_component_commit bridge-marker)
BRIDGE_MARKER_REPO=$(yaml-utils::get_component_repo ${BRIDGE_MARKER_URL})

TEMP_DIR=$(git-utils::create_temp_path bridge-marker)
trap "rm -rf ${TEMP_DIR}" EXIT
BRIDGE_MARKER_PATH=${TEMP_DIR}/${BRIDGE_MARKER_REPO}

echo 'Fetch bridge-marker sources'
git-utils::fetch_component ${BRIDGE_MARKER_PATH} ${BRIDGE_MARKER_URL} ${BRIDGE_MARKER_COMMIT}

echo 'Get bridge-marker image name and update it under CNAO'
BRIDGE_MARKER_TAG=$(git-utils::get_component_tag ${BRIDGE_MARKER_PATH})
BRIDGE_MARKER_IMAGE=quay.io/kubevirt/bridge-marker
BRIDGE_MARKER_IMAGE_TAGGED=${BRIDGE_MARKER_IMAGE}:${BRIDGE_MARKER_TAG}
sed -i "s#\"${BRIDGE_MARKER_IMAGE}:.*\"#\"${BRIDGE_MARKER_IMAGE_TAGGED}\"#" \
    pkg/components/components.go \
    test/releases/${CNAO_VERSION}.go

