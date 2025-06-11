#!/usr/bin/env bash

set -xeo pipefail

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh
source hack/components/docker-utils.sh

function __parametize_by_object() {
  for f in ./*; do
    case "${f}" in
      ./Namespace_secondary.yaml)
        yaml-utils::update_param ${f} metadata.name '{{ .Namespace }}'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
      ./ConfigMap_secondary-dns.yaml)
        yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
        yaml-utils::set_param ${f} data.DOMAIN '{{ .Domain }}'
        yaml-utils::set_param ${f} data.NAME_SERVER_IP '{{ .NameServerIp }}'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
      ./ClusterRoleBinding_secondary.yaml)
        yaml-utils::update_param ${f} subjects[0].namespace '{{ .Namespace }}'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
      ./Deployment_secondary-dns.yaml)
        yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
        yaml-utils::update_param ${f} spec.template.spec.containers[0].image '{{ .CoreDNSImage }}'
        yaml-utils::update_param ${f} spec.template.spec.securityContext.runAsNonRoot '{{ .RunAsNonRoot }}'
        yaml-utils::update_param ${f} spec.template.spec.securityContext.runAsUser '{{ .RunAsUser }}'
        yaml-utils::update_param ${f} spec.template.spec.containers[1].image '{{ .KubeSecondaryDNSImage }}'
        yaml-utils::set_param ${f} spec.template.spec.containers[0].imagePullPolicy '{{ .ImagePullPolicy }}'
        yaml-utils::set_param ${f} spec.template.spec.containers[1].imagePullPolicy '{{ .ImagePullPolicy }}'
        yaml-utils::set_param ${f} spec.template.spec.nodeSelector '{{ toYaml .Placement.NodeSelector | nindent 8 }}'
        yaml-utils::set_param ${f} spec.template.spec.affinity '{{ toYaml .Placement.Affinity | nindent 8 }}'
        yaml-utils::set_param ${f} spec.template.spec.tolerations '{{ toYaml .Placement.Tolerations | nindent 8 }}'
        yaml-utils::set_param ${f} 'spec.template.metadata.annotations."openshift.io/required-scc"' '"restricted-v2"'
        yaml-utils::set_param ${f} 'spec.template.metadata.labels."hco.kubevirt.io/allow-access-cluster-services"' '""'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
      ./ServiceAccount_secondary.yaml)
        yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
    esac
  done
}

echo 'Bumping kube-secondary-dns'
KUBE_SECONDARY_DNS_URL=$(yaml-utils::get_component_url kube-secondary-dns)
KUBE_SECONDARY_DNS_COMMIT=$(yaml-utils::get_component_commit kube-secondary-dns)
KUBE_SECONDARY_DNS_REPO=$(yaml-utils::get_component_repo ${KUBE_SECONDARY_DNS_URL})

TEMP_DIR=$(git-utils::create_temp_path kube-secondary-dns)
trap "rm -rf ${TEMP_DIR}" EXIT
KUBE_SECONDARY_DNS_PATH=${TEMP_DIR}/${KUBE_SECONDARY_DNS_REPO}

echo 'Fetch kube-secondary-dns sources'
git-utils::fetch_component ${KUBE_SECONDARY_DNS_PATH} ${KUBE_SECONDARY_DNS_URL} ${KUBE_SECONDARY_DNS_COMMIT}

echo 'Adjust kube-secondary-dns to CNAO'
(
  cd ${KUBE_SECONDARY_DNS_PATH}
  mkdir -p config/cnao
  cp manifests/secondarydns.yaml config/cnao

  echo 'Split manifest per object'
  cd config/cnao
  $(yaml-utils::split_yaml_by_seperator . secondarydns.yaml)

  rm secondarydns.yaml
  $(yaml-utils::rename_files_by_object .)

  echo 'parametize manifests by object'
  __parametize_by_object

  echo 'rejoin sub-manifests to a final manifest'
  cat Namespace_secondary.yaml \
      ConfigMap_secondary-dns.yaml \
      ClusterRole_secondary.yaml \
      ClusterRoleBinding_secondary.yaml \
      ServiceAccount_secondary.yaml \
      Deployment_secondary-dns.yaml > secondarydns.yaml
)

echo 'copy manifests'
rm -rf data/kube-secondary-dns/*
cp ${KUBE_SECONDARY_DNS_PATH}/config/cnao/secondarydns.yaml data/kube-secondary-dns

echo 'Get kube-secondary-dns image name and update it under CNAO'
KUBE_SECONDARY_DNS_TAG=$(git-utils::get_component_tag ${KUBE_SECONDARY_DNS_PATH})
KUBE_SECONDARY_DNS_IMAGE=ghcr.io/kubevirt/kubesecondarydns
KUBE_SECONDARY_DNS_IMAGE_TAGGED=${KUBE_SECONDARY_DNS_IMAGE}:${KUBE_SECONDARY_DNS_TAG}
KUBE_SECONDARY_DNS_IMAGE_DIGEST="$(docker-utils::get_image_digest "${KUBE_SECONDARY_DNS_IMAGE_TAGGED}" "${KUBE_SECONDARY_DNS_IMAGE}")"

CORE_DNS_IMAGE=registry.k8s.io/coredns/coredns
CORE_DNS_IMAGE_TAGGED=$(grep $CORE_DNS_IMAGE ${KUBE_SECONDARY_DNS_PATH}/manifests/secondarydns.yaml | awk -F": " '{print $2}')
CORE_DNS_IMAGE_DIGEST="$(docker-utils::get_image_digest "${CORE_DNS_IMAGE_TAGGED}" "${CORE_DNS_IMAGE}")"

sed -i -r "s#\"${KUBE_SECONDARY_DNS_IMAGE}(@sha256)?:.*\"#\"${KUBE_SECONDARY_DNS_IMAGE_DIGEST}\"#" pkg/components/components.go
sed -i -r "s#\"${KUBE_SECONDARY_DNS_IMAGE}(@sha256)?:.*\"#\"${KUBE_SECONDARY_DNS_IMAGE_DIGEST}\"#" test/releases/${CNAO_VERSION}.go

sed -i -r "s#\"${CORE_DNS_IMAGE}(@sha256)?:.*\"#\"${CORE_DNS_IMAGE_DIGEST}\"#" pkg/components/components.go
sed -i -r "s#\"${CORE_DNS_IMAGE}(@sha256)?:.*\"#\"${CORE_DNS_IMAGE_DIGEST}\"#" test/releases/${CNAO_VERSION}.go
