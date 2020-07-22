#!/usr/bin/env bash

set -xeo pipefail

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh

echo 'Bumping macvtap-cni'
MACVTAP_URL=$(yaml-utils::get_component_url macvtap-cni)
MACVTAP_COMMIT=$(yaml-utils::get_component_commit macvtap-cni)
MACVTAP_REPO=$(yaml-utils::get_component_repo ${MACVTAP_URL})

TEMP_DIR=$(git-utils::create_temp_path macvtap-cni)
trap "rm -rf ${TEMP_DIR}" EXIT
MACVTAP_PATH=${TEMP_DIR}/${MACVTAP_REPO}

echo 'Fetch macvtap-cni sources'
git-utils::fetch_component ${MACVTAP_PATH} ${MACVTAP_URL} ${MACVTAP_COMMIT}

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
MACVTAP_TAG=$(git-utils::get_component_tag ${MACVTAP_PATH})
MACVTAP_IMAGE=quay.io/kubevirt/macvtap-cni
MACVTAP_IMAGE_TAGGED=${MACVTAP_IMAGE}:${MACVTAP_TAG}
sed -i "s#\"${MACVTAP_IMAGE}:.*\"#\"${MACVTAP_IMAGE_TAGGED}\"#" pkg/components/components.go
# TODO: uncomment the following line *once* there is macvtap upgrade is supported
#sed -i "s#\"${MACVTAP_IMAGE}:.*\"#\"${MACVTAP_IMAGE_TAGGED}\"#" test/releases/${CNAO_VERSION}.go
