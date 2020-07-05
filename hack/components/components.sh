#!/usr/bin/env bash

set -xe

source hack/components/common_functions.sh

CNAO_VERSION=${VERSION} # Exported from Makefile

echo 'Setup temporary Go path'
export GOPATH=${PWD}/_components/go
mkdir -p $GOPATH
export PATH=${GOPATH}/bin:${PATH}

echo 'kubemacpool'
source hack/components/bump_kubemacpool.sh

echo 'macvtap'
MACVTAP_URL=$(cat components.yaml | shyaml get-value components.macvtap-cni.url)
MACVTAP_COMMIT=$(cat components.yaml | shyaml get-value components.macvtap-cni.commit)
MACVTAP_REPO=$(echo ${MACVTAP_URL} | sed 's#https://\(.*\)#\1#')
MACVTAP_PATH=${GOPATH}/src/${MACVTAP_REPO}

echo 'Fetch macvtap-cni sources'
fetch_component ${MACVTAP_PATH} ${MACVTAP_URL} ${MACVTAP_COMMIT}

rm -rf data/macvtap/*
echo 'Copy the templates from the macvtap-cni repo ...'
cp ${MACVTAP_PATH}/templates/namespace.yaml.in data/macvtap/000-ns.yaml
echo "{{ if .EnableSCC }}" >> data/macvtap/001-rbac.yaml
cat ${MACVTAP_PATH}/templates/scc.yaml.in >> data/macvtap/001-rbac.yaml
echo "{{ end }}" >> data/macvtap/001-rbac.yaml
cat <<EOF > data/macvtap/002-macvtap-daemonset.yaml
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: macvtap-deviceplugin-config
  namespace: {{ .Namespace }}
data:
  DP_MACVTAP_CONF: "[]"
---
EOF
cat ${MACVTAP_PATH}/templates/macvtap.yaml.in >> data/macvtap/002-macvtap-daemonset.yaml

echo 'Get macvtap-cni image name and update it under CNAO'
MACVTAP_TAG=$(get_component_tag ${MACVTAP_PATH})
MACVTAP_IMAGE=quay.io/kubevirt/macvtap-cni
MACVTAP_IMAGE_TAGGED=${MACVTAP_IMAGE}:${MACVTAP_TAG}
sed -i "s#\"${MACVTAP_IMAGE}:.*\"#\"${MACVTAP_IMAGE_TAGGED}\"#" pkg/components/components.go
# TODO: uncomment the following line *once* there is macvtap upgrade is supported
#sed -i "s#\"${MACVTAP_IMAGE}:.*\"#\"${MACVTAP_IMAGE_TAGGED}\"#" test/releases/${CNAO_VERSION}.go

echo 'Clone kubernetes-nmstate'
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
