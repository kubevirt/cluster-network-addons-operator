#!/usr/bin/env bash

set -xe

source hack/components/common_functions.sh

NMSTATE_URL=$(cat components.yaml | shyaml get-value components.nmstate.url)
NMSTATE_COMMIT=$(cat components.yaml | shyaml get-value components.nmstate.commit)
NMSTATE_REPO=$(echo ${NMSTATE_URL} | sed 's#https://\(.*\)#\1#')
NMSTATE_PATH=${GOPATH}/src/${NMSTATE_REPO}
fetch_component ${NMSTATE_PATH} ${NMSTATE_URL} ${NMSTATE_COMMIT}

echo 'Get kubernetes-nmstate image name and update it under CNAO'
NMSTATE_TAG=$(get_component_tag ${NMSTATE_PATH})
NMSTATE_IMAGE=quay.io/nmstate/kubernetes-nmstate-handler
NMSTATE_IMAGE_TAGGED=${NMSTATE_IMAGE}:${NMSTATE_TAG}
sed -i "s#\"${NMSTATE_IMAGE}:.*\"#\"${NMSTATE_IMAGE_TAGGED}\"#" \
    pkg/components/components.go \
    test/releases/${CNAO_VERSION}.go

echo 'Copy kubernetes-nmstate manifests'
rm -rf data/nmstate/*
cp $NMSTATE_PATH/deploy/handler/* data/nmstate/
cp $NMSTATE_PATH/deploy/crds/*nodenetwork*crd* data/nmstate/
