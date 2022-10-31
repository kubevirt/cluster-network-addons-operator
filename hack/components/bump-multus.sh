#!/usr/bin/env bash

set -xeo pipefail

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh
source hack/components/docker-utils.sh

#here we do all the object specific parametizing
function __parametize_by_object() {
	for f in ./*; do
		case "${f}" in
			./ClusterRoleBinding_multus.yaml)
				yaml-utils::update_param ${f} subjects[0].namespace '{{ .Namespace }}'
				yaml-utils::remove_single_quotes_from_yaml ${f}
				;;
			./ServiceAccount_multus.yaml)
				yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
				yaml-utils::remove_single_quotes_from_yaml ${f}
				;;
                       ./ConfigMap_multus-daemon-config.yaml)
                               yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
                               ;;
			./DaemonSet_kube-multus-ds.yaml)
				yaml-utils::update_param ${f} metadata.name 'multus'
				yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
				yaml-utils::update_param ${f} spec.selector.matchLabels.name 'kube-multus-ds-amd64'
				yaml-utils::update_param ${f} spec.template.metadata.labels.name 'kube-multus-ds-amd64'
				yaml-utils::update_param ${f} spec.template.spec.containers[0].image '{{ .MultusImage }}'
				yaml-utils::set_param ${f} spec.template.spec.containers[0].imagePullPolicy '{{ .ImagePullPolicy }}'
				yaml-utils::update_param ${f} spec.template.spec.initContainers[0].image '{{ .MultusImage }}'
				yaml-utils::set_param ${f} spec.template.spec.priorityClassName 'system-cluster-critical'
				yaml-utils::update_param ${f} spec.template.spec.volumes[0].hostPath.path '{{ .CNIConfigDir }}'
				yaml-utils::update_param ${f} spec.template.spec.volumes[1].hostPath.path '{{ .CNIBinDir }}'
				yaml-utils::delete_param ${f} spec.template.spec.containers[0].resources.limits
				yaml-utils::update_param ${f} spec.template.spec.containers[0].resources.requests.cpu '"10m"'
				yaml-utils::update_param ${f} spec.template.spec.containers[0].resources.requests.memory '"15Mi"'
				yaml-utils::set_param ${f} spec.template.spec.nodeSelector '{{ toYaml .Placement.NodeSelector | nindent 8 }}'
				yaml-utils::set_param ${f} spec.template.spec.containers[0].lifecycle.preStop.exec.command '["/bin/sh", "-c", "rm -rf /host/etc/cni/net.d/00-multus.conf /host/var/lib/cni/*"]'
				yaml-utils::set_param ${f} spec.template.spec.affinity '{{ toYaml .Placement.Affinity | nindent 8 }}'
				yaml-utils::update_param ${f} spec.template.spec.tolerations '{{ toYaml .Placement.Tolerations | nindent 8 }}'
				yaml-utils::remove_single_quotes_from_yaml ${f}
				;;
		esac
	done
}

echo 'Bumping multus'
MULTUS_URL=$(yaml-utils::get_component_url multus)
MULTUS_COMMIT=$(yaml-utils::get_component_commit multus)
MULTUS_REPO=$(yaml-utils::get_component_repo ${MULTUS_URL})

TEMP_DIR=$(git-utils::create_temp_path multus)
trap "rm -rf ${TEMP_DIR}" EXIT
MULTUS_PATH=${TEMP_DIR}/${MULTUS_REPO}

echo 'Fetch multus sources'
git-utils::fetch_component ${MULTUS_PATH} ${MULTUS_URL} ${MULTUS_COMMIT}

(
	cd ${MULTUS_PATH}
	mkdir -p config/cnao
	cp deployments/multus-daemonset-thick.yml config/cnao

	echo 'Split manifest per object'
	cd config/cnao
	grep -v "^#" multus-daemonset-thick.yml > no-comments-multus-daemonset-thick.yml
	$(yaml-utils::split_yaml_by_seperator . no-comments-multus-daemonset-thick.yml)
       rm multus-daemonset-thick.yml no-comments-multus-daemonset-thick.yml
	$(yaml-utils::rename_files_by_object .)

	echo 'parametize manifests by object'
	__parametize_by_object

	cat <<EOF > 000-ns.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Namespace }}
EOF

	cat <<EOF > SecurityContextConstraints_multus.yaml
{{ if .EnableSCC }}
---
apiVersion: security.openshift.io/v1
kind: SecurityContextConstraints
metadata:
  name: multus
allowPrivilegedContainer: true
allowHostDirVolumePlugin: true
readOnlyRootFilesystem: false
allowHostIPC: false
allowHostNetwork: true
allowHostPID: false
allowHostPorts: false
runAsUser:
  type: RunAsAny
seLinuxContext:
  type: RunAsAny
users:
- system:serviceaccount:{{ .Namespace }}:multus
volumes:
- "*"
{{ end }}
---
EOF

	echo 'rejoin sub-manifests to final manifest'
	YAML_FILE=001-multus.yaml
	touch ${YAML_FILE}
	cat CustomResourceDefinition_network-attachment-definitions.k8s.cni.cncf.io.yaml >> ${YAML_FILE} &&
		cat ClusterRole_multus.yaml >> ${YAML_FILE} &&
		cat ClusterRoleBinding_multus.yaml >> ${YAML_FILE} &&
		cat ServiceAccount_multus.yaml >> ${YAML_FILE} &&
		cat ConfigMap_multus-daemon-config.yaml >> ${YAML_FILE} &&
		cat DaemonSet_kube-multus-ds.yaml >> ${YAML_FILE} &&
		cat SecurityContextConstraints_multus.yaml >> ${YAML_FILE}
)

echo 'copy manifests'
rm -rf data/multus/*
cp ${MULTUS_PATH}/config/cnao/000-ns.yaml data/multus/
cp ${MULTUS_PATH}/config/cnao/001-multus.yaml data/multus/

echo 'Get multus image name'
MULTUS_TAG=$(git-utils::get_component_tag ${MULTUS_PATH})
MULTUS_IMAGE=ghcr.io/k8snetworkplumbingwg/multus-cni
# TODO: do not use this hardcoded image, and rely on tagged versions instead.
# had to resort to this shenanigan because of https://github.com/k8snetworkplumbingwg/multus-cni/issues/944
MULTUS_IMAGE_TAGGED=${MULTUS_IMAGE}@sha256:09a572e8bdf8a398db024ca252d06cf3ac0a03e07ae547d6a84221d4f6a9f96f
if [[ -n "$(docker-utils::check_image_exists "${MULTUS_IMAGE}" "${MULTUS_TAG}")" ]]; then
    MULTUS_IMAGE_DIGEST="$(docker-utils::get_image_digest "${MULTUS_IMAGE_TAGGED}" "${MULTUS_IMAGE}")"
else
    MULTUS_IMAGE_DIGEST=${MULTUS_IMAGE_TAGGED}
fi

echo 'Update multus references under CNAO'
sed -i -r "s#\"${MULTUS_IMAGE}(@sha256)?:.*\"#\"${MULTUS_IMAGE_DIGEST}\"#" pkg/components/components.go
sed -i -r "s#\"${MULTUS_IMAGE}(@sha256)?:.*\"#\"${MULTUS_IMAGE_DIGEST}\"#" test/releases/${CNAO_VERSION}.go
