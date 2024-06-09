#!/usr/bin/env bash

set -xeo pipefail

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh
source hack/components/docker-utils.sh

IPAMCLAIMS_CRD_VERSION="v0.4.0-alpha"

function __parametize_by_object() {
  for f in ./*; do
    case "${f}" in
      ./Namespace_kubevirt-ipam-claims-system.yaml)
        yaml-utils::update_param ${f} metadata.name '{{ .Namespace }}'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
      ./ClusterRoleBinding_kubevirt-ipam-claims-manager-rolebinding.yaml)
        yaml-utils::update_param ${f} subjects[0].namespace '{{ .Namespace }}'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
      ./Deployment_kubevirt-ipam-claims-controller-manager.yaml)
        yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
        yaml-utils::update_param ${f} spec.template.spec.containers[0].image '{{ .KubevirtIpamControllerImage }}'
        yaml-utils::set_param ${f} spec.template.spec.containers[0].imagePullPolicy '{{ .ImagePullPolicy }}'
        yaml-utils::set_param ${f} spec.template.spec.containers[0].args[1] '"--certificates-dir={{ .CertDir }}"'
        yaml-utils::set_param ${f} spec.template.spec.containers[0].volumeMounts[0].mountPath '{{ .MountPath }}'
        yaml-utils::set_param ${f} spec.template.spec.volumes[0].secret.secretName '{{ .SecretName }}'
        yaml-utils::set_param ${f} spec.template.spec.nodeSelector '{{ toYaml .Placement.NodeSelector | nindent 8 }}'
        yaml-utils::set_param ${f} spec.template.spec.affinity '{{ toYaml .Placement.Affinity | nindent 8 }}'
        yaml-utils::set_param ${f} spec.template.spec.tolerations '{{ toYaml .Placement.Tolerations | nindent 8 }}'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
      ./Service_kubevirt-ipam-claims-webhook-service.yaml)
        yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
      ./Certificate_kubevirt-ipam-claims-serving-cert.yaml)
        yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
        yaml-utils::update_param ${f} spec.dnsNames[0] 'kubevirt-ipam-claims-webhook-service.{{ .Namespace }}.svc'
        yaml-utils::update_param ${f} spec.dnsNames[1] 'kubevirt-ipam-claims-webhook-service.{{ .Namespace }}.svc.cluster.local'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
      ./Issuer_kubevirt-ipam-claims-selfsigned-issuer.yaml)
        yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
      ./MutatingWebhookConfiguration_kubevirt-ipam-claims-mutating-webhook-configuration.yaml)
        yaml-utils::update_param ${f} webhooks[0].clientConfig.service.namespace '{{ .Namespace }}'
        sed -i '/cert-manager.io\/inject-ca-from/c\    {{ .WebhookAnnotation }}' ${f}
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
      ./RoleBinding_kubevirt-ipam-claims-leader-election-rolebinding.yaml)
        yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
        yaml-utils::update_param ${f} subjects[0].namespace '{{ .Namespace }}'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
      ./Role_kubevirt-ipam-claims-leader-election-role.yaml)
        yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
      ./ServiceAccount_kubevirt-ipam-claims-controller-manager.yaml)
        yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
      ./Service_kubevirt-ipam-claims-webhook-service.yaml)
        yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
        yaml-utils::remove_single_quotes_from_yaml ${f}
        ;;
    esac
  done
}

echo 'Bumping kubevirt-ipam-controller'
KUBEVIRT_IPAM_CONTROLLER_URL=$(yaml-utils::get_component_url kubevirt-ipam-controller)
KUBEVIRT_IPAM_CONTROLLER_COMMIT=$(yaml-utils::get_component_commit kubevirt-ipam-controller)
KUBEVIRT_IPAM_CONTROLLER_REPO=$(yaml-utils::get_component_repo ${KUBEVIRT_IPAM_CONTROLLER_URL})

TEMP_DIR=$(git-utils::create_temp_path kubevirt-ipam-controller)
trap "rm -rf ${TEMP_DIR}" EXIT
KUBEVIRT_IPAM_CONTROLLER_PATH=${TEMP_DIR}/${KUBEVIRT_IPAM_CONTROLLER_REPO}

echo 'Fetch kubevirt-ipam-controller sources'
git-utils::fetch_component ${KUBEVIRT_IPAM_CONTROLLER_PATH} ${KUBEVIRT_IPAM_CONTROLLER_URL} ${KUBEVIRT_IPAM_CONTROLLER_COMMIT}

echo 'Adjust kubevirt-ipam-controller to CNAO'
(
  cd ${KUBEVIRT_IPAM_CONTROLLER_PATH}
  mkdir -p config/cnao
  cp dist/install.yaml config/cnao

  echo 'Split manifest per object'
  cd config/cnao

  $(yaml-utils::split_yaml_by_seperator . install.yaml)

  rm install.yaml
  $(yaml-utils::rename_files_by_object .)

  echo 'parametize manifests by object'
  __parametize_by_object

  sed -i '1i{{ if not .IsOpenshift }}' Issuer_kubevirt-ipam-claims-selfsigned-issuer.yaml
  echo "{{ end }}" >> Issuer_kubevirt-ipam-claims-selfsigned-issuer.yaml

  sed -i '1i{{ if not .IsOpenshift }}' Certificate_kubevirt-ipam-claims-serving-cert.yaml
  echo "{{ end }}" >> Certificate_kubevirt-ipam-claims-serving-cert.yaml

  sed -i '/metadata:/a\{{ if .IsOpenshift }}\
  annotations:\
    service.beta.openshift.io/serving-cert-secret-name: kubevirt-ipam-claims-webhook-service\
{{ end }}' Service_kubevirt-ipam-claims-webhook-service.yaml

  echo 'rejoin sub-manifests to a final manifest'
  cat Namespace_kubevirt-ipam-claims-system.yaml \
      ServiceAccount_kubevirt-ipam-claims-controller-manager.yaml \
      Role_kubevirt-ipam-claims-leader-election-role.yaml \
      ClusterRole_kubevirt-ipam-claims-manager-role.yaml \
      RoleBinding_kubevirt-ipam-claims-leader-election-rolebinding.yaml \
      ClusterRoleBinding_kubevirt-ipam-claims-manager-rolebinding.yaml \
      Service_kubevirt-ipam-claims-webhook-service.yaml \
      Deployment_kubevirt-ipam-claims-controller-manager.yaml \
      Certificate_kubevirt-ipam-claims-serving-cert.yaml \
      Issuer_kubevirt-ipam-claims-selfsigned-issuer.yaml \
      MutatingWebhookConfiguration_kubevirt-ipam-claims-mutating-webhook-configuration.yaml > 001-kubevirtipamcontroller.yaml

)

echo 'Copy manifests'
rm -rf data/kubevirt-ipam-controller/*

# CRD
crd_manifest="https://raw.githubusercontent.com/k8snetworkplumbingwg/ipamclaims/${IPAMCLAIMS_CRD_VERSION}/artifacts/k8s.cni.cncf.io_ipamclaims.yaml"
echo "{{ if not .IsOpenshift }}" > data/kubevirt-ipam-controller/000-crd.yaml
curl $crd_manifest >> data/kubevirt-ipam-controller/000-crd.yaml
echo "{{ end }}" >> data/kubevirt-ipam-controller/000-crd.yaml

# Kubevirt Ipam controller
cp ${KUBEVIRT_IPAM_CONTROLLER_PATH}/config/cnao/001-kubevirtipamcontroller.yaml data/kubevirt-ipam-controller
sed -i '/app\.kubernetes\.io\//d' data/kubevirt-ipam-controller/001-kubevirtipamcontroller.yaml

echo 'Get kubevirt-ipam-controller image name and update it under CNAO'
KUBEVIRT_IPAM_CONTROLLER_TAG=$(git-utils::get_component_tag ${KUBEVIRT_IPAM_CONTROLLER_PATH})
KUBEVIRT_IPAM_CONTROLLER_IMAGE=ghcr.io/maiqueb/kubevirt-ipam-claims
KUBEVIRT_IPAM_CONTROLLER_IMAGE_TAGGED=${KUBEVIRT_IPAM_CONTROLLER_IMAGE}:${KUBEVIRT_IPAM_CONTROLLER_TAG}
KUBEVIRT_IPAM_CONTROLLER_IMAGE_DIGEST="$(docker-utils::get_image_digest "${KUBEVIRT_IPAM_CONTROLLER_IMAGE_TAGGED}" "${KUBEVIRT_IPAM_CONTROLLER_IMAGE}")"

sed -i -r "s#\"${KUBEVIRT_IPAM_CONTROLLER_IMAGE}(@sha256)?:.*\"#\"${KUBEVIRT_IPAM_CONTROLLER_IMAGE_DIGEST}\"#" pkg/components/components.go
sed -i -r "s#\"${KUBEVIRT_IPAM_CONTROLLER_IMAGE}(@sha256)?:.*\"#\"${KUBEVIRT_IPAM_CONTROLLER_IMAGE_DIGEST}\"#" test/releases/${CNAO_VERSION}.go
