#!/usr/bin/env bash

set -xeo pipefail

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh

echo 'Bumping multus'
MULTUS_URL=$(yaml-utils::get_component_url multus)
MULTUS_COMMIT=$(yaml-utils::get_component_commit multus)
MULTUS_REPO=$(yaml-utils::get_component_repo ${MULTUS_URL})

TEMP_DIR=$(git-utils::create_temp_path multus)
trap "rm -rf ${TEMP_DIR}" EXIT
MULTUS_PATH=${TEMP_DIR}/${MULTUS_REPO}

echo 'Fetch multus sources'
git-utils::fetch_component ${MULTUS_PATH} ${MULTUS_URL} ${MULTUS_COMMIT}

echo 'Get multus image name and update it under CNAO'
MULTUS_TAG=$(git-utils::get_component_tag ${MULTUS_PATH})
MULTUS_IMAGE=nfvpe/multus
MULTUS_IMAGE_TAGGED=${MULTUS_IMAGE}:${MULTUS_TAG}
sed -i "s#\"${MULTUS_IMAGE}:.*\"#\"${MULTUS_IMAGE_TAGGED}\"#" \
    pkg/components/components.go \
    test/releases/${CNAO_VERSION}.go
