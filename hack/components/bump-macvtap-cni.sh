#!/usr/bin/env bash

set -xeo pipefail

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh
source hack/components/docker-utils.sh

function __parametize_by_object() {
  for f in ./*; do
    case "${f}" in
      ./macvtap.yaml.in)
        yaml-utils::set_param ${f} 'spec.template.spec.serviceAccountName' 'macvtap-cni'
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
cat <<'EOF' > data/macvtap/001-rbac.yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: macvtap-cni
  namespace: {{ .Namespace }}
{{ if .EnableSCC }}
---
apiVersion: security.openshift.io/v1
kind: SecurityContextConstraints
metadata:
  name: macvtap-cni
allowHostNetwork: true
allowPrivilegedContainer: true
allowHostDirVolumePlugin: true
allowHostIPC: false
allowHostPID: false
allowHostPorts: false
readOnlyRootFilesystem: false
runAsUser:
  type: RunAsAny
seLinuxContext:
  type: RunAsAny
volumes:
  - hostPath
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: macvtap-cni-scc-use
rules:
- apiGroups: ["security.openshift.io"]
  resources: ["securitycontextconstraints"]
  resourceNames: ["macvtap-cni"]
  verbs: ["use"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: macvtap-cni-scc-use
  namespace: {{ .Namespace }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: macvtap-cni-scc-use
subjects:
- kind: ServiceAccount
  name: macvtap-cni
  namespace: {{ .Namespace }}
{{ end }}
EOF
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
