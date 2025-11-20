#!/usr/bin/env bash

set -xeo pipefail

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh
source hack/components/docker-utils.sh

#here we do all the object specific parametizing
function __parametize_by_object() {
	for f in ./*; do
		case "${f}" in
			./ClusterRoleBinding_bridge-marker-crb.yaml)
				yaml-utils::update_param ${f} subjects[0].namespace  '{{ .Namespace }}'
				yaml-utils::remove_single_quotes_from_yaml ${f}
				;;
			./ServiceAccount_bridge-marker.yaml)
				yaml-utils::update_param ${f} metadata.namespace  '{{ .Namespace }}'
				yaml-utils::remove_single_quotes_from_yaml ${f}
				;;
			./DaemonSet_bridge-marker.yaml)
				yaml-utils::update_param ${f} metadata.namespace  '{{ .Namespace }}'
				yaml-utils::update_param ${f} spec.template.spec.containers[0].image  '{{ .LinuxBridgeMarkerImage }}'
				yaml-utils::update_param ${f} spec.template.spec.containers[0].imagePullPolicy  '{{ .ImagePullPolicy }}'
				yaml-utils::update_param ${f} spec.template.spec.nodeSelector '{{ toYaml .Placement.NodeSelector | nindent 8 }}'
				yaml-utils::set_param ${f} spec.template.spec.affinity '{{ toYaml .Placement.Affinity | nindent 8 }}'
				yaml-utils::set_param ${f} 'spec.template.metadata.annotations."openshift.io/required-scc"' '"bridge-marker"'
				yaml-utils::update_param ${f} spec.template.spec.tolerations '{{ toYaml .Placement.Tolerations | nindent 8 }}'
				yaml-utils::set_param ${f} spec.template.spec.volumes[0].name 'tmp'
				yaml-utils::set_param ${f} spec.template.spec.volumes[0].emptyDir '{}'
				yaml-utils::set_param ${f} spec.template.spec.containers[0].volumeMounts[0].name 'tmp'
				yaml-utils::set_param ${f} spec.template.spec.containers[0].volumeMounts[0].mountPath '/tmp'
				yaml-utils::set_param ${f} spec.template.spec.securityContext.runAsNonRoot 'true'
				yaml-utils::set_param ${f} spec.template.spec.securityContext.runAsUser '1001'
				yaml-utils::set_param ${f} spec.template.spec.containers[0].securityContext.readOnlyRootFilesystem 'true'
				yaml-utils::remove_single_quotes_from_yaml ${f}
				;;
		esac
	done
}

echo 'Bumping bridge-marker'
BRIDGE_MARKER_URL=$(yaml-utils::get_component_url bridge-marker)
BRIDGE_MARKER_COMMIT=$(yaml-utils::get_component_commit bridge-marker)
BRIDGE_MARKER_REPO=$(yaml-utils::get_component_repo ${BRIDGE_MARKER_URL})

TEMP_DIR=$(git-utils::create_temp_path bridge-marker)
trap "rm -rf ${TEMP_DIR}" EXIT
BRIDGE_MARKER_PATH=${TEMP_DIR}/${BRIDGE_MARKER_REPO}

echo 'Fetch bridge-marker sources'
git-utils::fetch_component ${BRIDGE_MARKER_PATH} ${BRIDGE_MARKER_URL} ${BRIDGE_MARKER_COMMIT}

(
	cd ${BRIDGE_MARKER_PATH}
	mkdir -p config/cnao
	cp manifests/bridge-marker.yml.in config/cnao

	echo 'Split manifest per object'
	cd config/cnao
	#in order for it to split properly it needs to start with ---
	$(yaml-utils::append_delimiter bridge-marker.yml.in)
	$(yaml-utils::split_yaml_by_seperator . bridge-marker.yml.in)
	rm bridge-marker.yml.in
	$(yaml-utils::rename_files_by_object .)

	echo 'parametize manifests by object'
	__parametize_by_object

	cat <<EOF > SecurityContextConstraints_bridge-marker.yaml
{{ if .EnableSCC }}
---
apiVersion: security.openshift.io/v1
kind: SecurityContextConstraints
metadata:
  name: bridge-marker
allowHostNetwork: true
allowHostDirVolumePlugin: false
allowPrivilegedContainer: false
readOnlyRootFilesystem: true
allowHostIPC: false
allowHostPID: false
allowHostPorts: false
runAsUser:
  type: MustRunAsNonRoot
seLinuxContext:
  type: MustRunAs
users:
- system:serviceaccount:{{ .Namespace }}:bridge-marker
volumes:
- configMap
- emptyDir
- projected
{{ end }}
---
EOF

	echo 'rejoin sub-manifests to final manifest'
	YAML_FILE=003-bridge-marker.yaml
	touch ${YAML_FILE}
	cat DaemonSet_bridge-marker.yaml >> ${YAML_FILE} &&
		cat ClusterRole_bridge-marker-cr.yaml >> ${YAML_FILE} &&
		cat ClusterRoleBinding_bridge-marker-crb.yaml >> ${YAML_FILE} &&
		cat ServiceAccount_bridge-marker.yaml >> ${YAML_FILE} &&
		cat SecurityContextConstraints_bridge-marker.yaml >> ${YAML_FILE}
)

echo 'copy manifests'
rm -f data/linux-bridge/003*.yaml
rm -f data/linux-bridge/0004*.yaml #remove old file
cp ${BRIDGE_MARKER_PATH}/config/cnao/003-bridge-marker.yaml data/linux-bridge/

echo 'Get bridge-marker image name and update it under CNAO'
BRIDGE_MARKER_TAG=$(git-utils::get_component_tag ${BRIDGE_MARKER_PATH})
BRIDGE_MARKER_IMAGE=quay.io/kubevirt/bridge-marker
BRIDGE_MARKER_IMAGE_TAGGED=${BRIDGE_MARKER_IMAGE}:${BRIDGE_MARKER_TAG}
BRIDGE_MARKER_IMAGE_DIGEST="$(docker-utils::get_image_digest "${BRIDGE_MARKER_IMAGE_TAGGED}" "${BRIDGE_MARKER_IMAGE}")"

sed -i -r "s#\"${BRIDGE_MARKER_IMAGE}(@sha256)?:.*\"#\"${BRIDGE_MARKER_IMAGE_DIGEST}\"#" pkg/components/components.go
sed -i -r "s#\"${BRIDGE_MARKER_IMAGE}(@sha256)?:.*\"#\"${BRIDGE_MARKER_IMAGE_DIGEST}\"#" test/releases/${CNAO_VERSION}.go
