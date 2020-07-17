#!/usr/bin/env bash

set -xeo pipefail

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh

echo Bumping kubernetes-nmstate
NMSTATE_URL=$(yaml-utils::get_component_url nmstate)
NMSTATE_COMMIT=$(yaml-utils::get_component_commit nmstate)
NMSTATE_REPO=$(yaml-utils::get_component_repo ${NMSTATE_URL})

TEMP_DIR=$(git-utils::create_temp_path nmstate)
trap "rm -rf ${TEMP_DIR}" EXIT
NMSTATE_PATH=${TEMP_DIR}/${NMSTATE_REPO}

echo 'Fetch kubernetes-nmstate sources'
git-utils::fetch_component ${NMSTATE_PATH} ${NMSTATE_URL} ${NMSTATE_COMMIT}

echo 'Get kubernetes-nmstate image name and update it under CNAO'
NMSTATE_TAG=$(git-utils::get_component_tag ${NMSTATE_PATH})
NMSTATE_IMAGE=quay.io/nmstate/kubernetes-nmstate-handler
NMSTATE_IMAGE_TAGGED=${NMSTATE_IMAGE}:${NMSTATE_TAG}
sed -i "s#\"${NMSTATE_IMAGE}:.*\"#\"${NMSTATE_IMAGE_TAGGED}\"#" \
    pkg/components/components.go \
    test/releases/${CNAO_VERSION}.go

echo 'Copy kubernetes-nmstate manifests'
rm -rf data/nmstate/*
cp $NMSTATE_PATH/deploy/handler/* data/nmstate/
cp $NMSTATE_PATH/deploy/crds/*nodenetwork*crd* data/nmstate/

echo 'Apply custom CNAO patches on kubernetes-nmstate manifests'
echo '
241a242,243
>   annotations:
>     networkaddonsoperator.network.kubevirt.io/rejectOwner: ""
' | patch data/nmstate/operator.yaml
