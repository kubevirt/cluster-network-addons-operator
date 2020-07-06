#!/usr/bin/env bash

set -xe

source hack/components/common-functions.sh

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
