#!/usr/bin/env bash

set -xeo pipefail

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh
source hack/components/docker-utils.sh

#here we do all the object specific parametizing
function __parametize_by_object() {
	for f in ./*; do
		case "${f}" in
			./ClusterRoleBinding_ovs-cni-marker-crb.yaml)
				yaml-utils::update_param ${f} subjects[0].namespace  '{{ .Namespace }}'
				yaml-utils::remove_single_quotes_from_yaml ${f}
				;;
			./ServiceAccount_ovs-cni-marker.yaml)
				yaml-utils::update_param ${f} metadata.namespace  '{{ .Namespace }}'
				yaml-utils::remove_single_quotes_from_yaml ${f}
				;;
			./DaemonSet_ovs-cni-amd64.yaml)
				yaml-utils::update_param ${f} metadata.namespace  '{{ .Namespace }}'
				yaml-utils::update_param ${f} spec.template.spec.initContainers[0].image  '{{ .OvsCNIImage }}'
				yaml-utils::update_param ${f} spec.template.spec.initContainers[0].imagePullPolicy  '{{ .ImagePullPolicy }}'
				yaml-utils::update_param ${f} spec.template.spec.initContainers[0].volumeMounts[0].mountPath  '/host/opt/cni/bin'
				yaml-utils::update_param ${f} spec.template.spec.containers[0].image  '{{ .OvsCNIImage }}'
				yaml-utils::update_param ${f} spec.template.spec.containers[0].imagePullPolicy  '{{ .ImagePullPolicy }}'
				yaml-utils::update_param ${f} spec.template.spec.volumes[0].hostPath.path  '{{ .CNIBinDir }}'
				yaml-utils::update_param ${f} spec.template.spec.nodeSelector '{{ toYaml .Placement.NodeSelector | nindent 8 }}'
				yaml-utils::set_param ${f} spec.template.spec.affinity '{{ toYaml .Placement.Affinity | nindent 8 }}'
				yaml-utils::update_param ${f} spec.template.spec.tolerations '{{ toYaml .Placement.Tolerations | nindent 8 }}'
				yaml-utils::remove_single_quotes_from_yaml ${f}
				;;
		esac
	done
}

echo 'Bumping ovs-cni'
OVS_URL=$(yaml-utils::get_component_url ovs-cni)
OVS_COMMIT=$(yaml-utils::get_component_commit ovs-cni)
OVS_REPO=$(yaml-utils::get_component_repo ${OVS_URL})

TEMP_DIR=$(git-utils::create_temp_path ovs-cni)
trap "rm -rf ${TEMP_DIR}" EXIT
OVS_PATH=${TEMP_DIR}/${OVS_REPO}

echo 'Fetch ovs-cni sources'
git-utils::fetch_component ${OVS_PATH} ${OVS_URL} ${OVS_COMMIT}

(
	cd ${OVS_PATH}
	mkdir -p config/cnao
	cp examples/ovs-cni.yml config/cnao

	echo 'Split manifest per object'
	cd config/cnao
	#in order for it to split properly it needs to start with ---
	$(yaml-utils::append_delimiter ovs-cni.yml)
	$(yaml-utils::split_yaml_by_seperator . ovs-cni.yml)
	rm ovs-cni.yml
	$(yaml-utils::rename_files_by_object .)

	echo 'parametize manifests by object'
	__parametize_by_object

	cat <<EOF > 000-ns.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Namespace }}
EOF

	cat <<EOF > SecurityContextConstraints_ovs-cni-marker.yaml
{{ if .EnableSCC }}
---
apiVersion: security.openshift.io/v1
kind: SecurityContextConstraints
metadata:
  name: ovs-cni-marker
allowHostNetwork: true
allowPrivilegedContainer: true
allowHostDirVolumePlugin: true
runAsUser:
  type: RunAsAny
seLinuxContext:
  type: RunAsAny
users:
  - system:serviceaccount:{{ .Namespace }}:ovs-cni-marker
{{ end }}
---
EOF

		echo 'rejoin sub-manifests to final manifest'
	YAML_FILE=001-ovs-cni.yaml
	touch ${YAML_FILE}
	cat DaemonSet_ovs-cni-amd64.yaml >> ${YAML_FILE} &&
		cat ClusterRole_ovs-cni-marker-cr.yaml >> ${YAML_FILE} &&
		cat ClusterRoleBinding_ovs-cni-marker-crb.yaml >> ${YAML_FILE} &&
		cat ServiceAccount_ovs-cni-marker.yaml >> ${YAML_FILE} &&
		cat SecurityContextConstraints_ovs-cni-marker.yaml >> ${YAML_FILE}
)

echo 'copy manifests'
rm -rf data/ovs/*
cp ${OVS_PATH}/config/cnao/000-ns.yaml data/ovs/
cp ${OVS_PATH}/config/cnao/001-ovs-cni.yaml data/ovs/

OVS_TAG=$(git-utils::get_component_tag ${OVS_PATH})

echo 'Get ovs-cni-plugin image name and update it under CNAO'
OVS_PLUGIN_IMAGE=quay.io/kubevirt/ovs-cni-plugin
OVS_PLUGIN_IMAGE_TAGGED=${OVS_PLUGIN_IMAGE}:${OVS_TAG}
OVS_PLUGIN_IMAGE_DIGEST="$(docker-utils::get_image_digest "${OVS_PLUGIN_IMAGE_TAGGED}" "${OVS_PLUGIN_IMAGE}")"

sed -i -r "s#\"${OVS_PLUGIN_IMAGE}(@sha256)?:.*\"#\"${OVS_PLUGIN_IMAGE_DIGEST}\"#" pkg/components/components.go
sed -i -r "s#\"${OVS_PLUGIN_IMAGE}(@sha256)?:.*\"#\"${OVS_PLUGIN_IMAGE_DIGEST}\"#" test/releases/${CNAO_VERSION}.go
