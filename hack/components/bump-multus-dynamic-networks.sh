#!/usr/bin/env bash

set -xeo pipefail

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh
source hack/components/docker-utils.sh

function __parametize_by_object() {
  for f in ./*; do
    case "${f}" in
      ./ConfigMap_dynamic-networks-controller-config.yaml)
        yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
        json_content=$(yaml-utils::get_param ${f} 'data."dynamic-networks-config.json"')
        updated_json=$(echo "${json_content}" | sed -E "s|\"criSocketPath\": *\"[^\"]*\"|\"criSocketPath\": \"/host{{ .HostCRISocketPath }}\"|")
        yaml-utils::set_param ${f} 'data."dynamic-networks-config.json"' "${updated_json}"$'\n'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
      ./ClusterRoleBinding_dynamic-networks-controller.yaml)
        yaml-utils::update_param ${f} subjects[0].namespace '{{ .Namespace }}'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
      ./DaemonSet_dynamic-networks-controller-ds.yaml)
        yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
        yaml-utils::set_param ${f} spec.template.spec.containers[0].imagePullPolicy '{{ .ImagePullPolicy }}'
        yaml-utils::update_param ${f} spec.template.spec.containers[0].image  '{{ .MultusDynamicNetworksControllerImage }}'
        yaml-utils::update_param ${f} spec.template.spec.containers[0].volumeMounts\(name=="cri-socket"\).mountPath '/host{{ .HostCRISocketPath }}'
        yaml-utils::set_param ${f} spec.template.spec.affinity '{{ toYaml .Placement.Affinity | nindent 8 }}'
        yaml-utils::update_param ${f} spec.template.spec.tolerations '{{ toYaml .Placement.Tolerations | nindent 8 }}'
        yaml-utils::update_param ${f} spec.template.spec.volumes\(name=="cri-socket"\).hostPath.path  '{{ .HostCRISocketPath }}'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
      ./ServiceAccount_dynamic-networks-controller.yaml)
        yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
    esac
  done
}

echo 'Bumping multus-dynamic-networks-controller'
MULTUS_DYNAMIC_NETWORKS_URL=$(yaml-utils::get_component_url multus-dynamic-networks)
MULTUS_DYNAMIC_NETWORKS_COMMIT=$(yaml-utils::get_component_commit multus-dynamic-networks)
MULTUS_DYNAMIC_NETWORKS_REPO=$(yaml-utils::get_component_repo ${MULTUS_DYNAMIC_NETWORKS_URL})

TEMP_DIR=$(git-utils::create_temp_path multus-dynamic-networks)
trap "rm -rf ${TEMP_DIR}" EXIT
MULTUS_DYNAMIC_NETWORKS_PATH=${TEMP_DIR}/${MULTUS_DYNAMIC_NETWORKS_REPO}

echo 'Fetch multus-dynamic-networks-controller sources'
git-utils::fetch_component ${MULTUS_DYNAMIC_NETWORKS_PATH} ${MULTUS_DYNAMIC_NETWORKS_URL} ${MULTUS_DYNAMIC_NETWORKS_COMMIT}

echo 'Adjust multus-dynamic-networks-controller to CNAO'
(
  cd ${MULTUS_DYNAMIC_NETWORKS_PATH}
  mkdir -p config/cnao
  cp manifests/crio-dynamic-networks-controller.yaml config/cnao

  echo 'Split manifest per object'
  cd config/cnao
  $(yaml-utils::split_yaml_by_seperator . crio-dynamic-networks-controller.yaml)

  rm crio-dynamic-networks-controller.yaml
  $(yaml-utils::rename_files_by_object .)

  echo 'parametize manifests by object'
  __parametize_by_object

  echo 'rejoin sub-manifests to final manifest'
  cat * > multus-dynamic-networks-controller.yaml
)

echo 'copy manifests'
rm -rf data/multus-dynamic-networks-controller/*
cp ${MULTUS_DYNAMIC_NETWORKS_PATH}/config/cnao/multus-dynamic-networks-controller.yaml data/multus-dynamic-networks-controller/000-controller.yaml

echo 'Get multus-dynamic-networks image name and update it under CNAO'
MULTUS_DYNAMIC_NETWORKS_TAG=$(git-utils::get_component_tag ${MULTUS_DYNAMIC_NETWORKS_PATH})
MULTUS_DYNAMIC_NETWORKS_IMAGE=ghcr.io/k8snetworkplumbingwg/multus-dynamic-networks-controller
MULTUS_DYNAMIC_NETWORKS_IMAGE_TAGGED=${MULTUS_DYNAMIC_NETWORKS_IMAGE}:${MULTUS_DYNAMIC_NETWORKS_TAG}
MULTUS_DYNAMIC_NETWORKS_IMAGE_DIGEST="$(docker-utils::get_image_digest "${MULTUS_DYNAMIC_NETWORKS_IMAGE_TAGGED}" "${MULTUS_DYNAMIC_NETWORKS_IMAGE}")"

sed -i -r "s#\"${MULTUS_DYNAMIC_NETWORKS_IMAGE}(@sha256)?:.*\"#\"${MULTUS_DYNAMIC_NETWORKS_IMAGE_DIGEST}\"#" pkg/components/components.go
sed -i -r "s#\"${MULTUS_DYNAMIC_NETWORKS_IMAGE}(@sha256)?:.*\"#\"${MULTUS_DYNAMIC_NETWORKS_IMAGE_DIGEST}\"#" test/releases/${CNAO_VERSION}.go
