#!/usr/bin/env bash

set -xeo pipefail

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh
source hack/components/docker-utils.sh

function __parametize_by_object() {
  for f in ./*; do
    case "${f}" in
      ./macvtap.yaml.in)
        yaml-utils::set_param ${f} spec.template.spec.affinity '{{ toYaml .Placement.Affinity | nindent 8 }}'
        yaml-utils::set_param ${f} spec.template.spec.nodeSelector '{{ toYaml .Placement.NodeSelector | nindent 8 }}'
        yaml-utils::set_param ${f} spec.template.spec.tolerations '{{ toYaml .Placement.Tolerations | nindent 8 }}'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
    esac
  done
}

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

(
  cd ${MACVTAP_PATH}/templates/
  echo 'parametize manifests by object'
  __parametize_by_object
)

cat ${MACVTAP_PATH}/templates/macvtap.yaml.in >> data/macvtap/002-macvtap-daemonset.yaml

echo 'Get macvtap-cni image name and update it under CNAO'
MACVTAP_TAG=$(git-utils::get_component_tag ${MACVTAP_PATH})
MACVTAP_IMAGE=quay.io/kubevirt/macvtap-cni
MACVTAP_IMAGE_TAGGED=${MACVTAP_IMAGE}:${MACVTAP_TAG}
MACVTAP_IMAGE_DIGEST="$(docker-utils::get_image_digest "${MACVTAP_IMAGE_TAGGED}" "${MACVTAP_IMAGE}")"

sed -i -r "s#\"${MACVTAP_IMAGE}(@sha256)?:.*\"#\"${MACVTAP_IMAGE_DIGEST}\"#" pkg/components/components.go
# TODO: uncomment the following line *once* there is macvtap upgrade is supported
#sed -i "s#\"${MACVTAP_IMAGE}(@sha256)?:.*\"#\"${MACVTAP_IMAGE_DIGEST}\"#" test/releases/${CNAO_VERSION}.go
