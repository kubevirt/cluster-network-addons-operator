#!/usr/bin/env bash

set -xeo pipefail

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh
source hack/components/docker-utils.sh

function __unquote_toYaml() {
	local yaml_file=$1
	sed -i "s/'{{ toYaml \(.*\)}}'/{{ toYaml \1}}/g" ${yaml_file}
}

function __unquote_numeric_and_boolean() {
	local yaml_file=$1
	sed -i "s/'{{ \.RunAsNonRoot }}'/{{ .RunAsNonRoot }}/g" ${yaml_file}
	sed -i "s/'{{ \.RunAsUser }}'/{{ .RunAsUser }}/g" ${yaml_file}
}

function __set_empty_string_label() {
	local yaml_file=$1
	local label_path=$2
	local label_name=${label_path##*.}
	label_name=${label_name//\"/}
	yaml-utils::set_param ${yaml_file} "${label_path}" ''
	sed -i "s/${label_name}:$/${label_name}: \"\"/" ${yaml_file}
}

function __parametize_by_object() {
	for f in ./*; do
		case "${f}" in
			./Namespace_kubemacpool-system.yaml)
				yaml-utils::update_param ${f} metadata.name '{{ .Namespace }}'
				yaml-utils::delete_param ${f} metadata.labels
				;;
			./ServiceAccount_kubemacpool-sa.yaml)
				yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
				;;
			./ClusterRoleBinding_kubemacpool-manager-rolebinding.yaml)
				yaml-utils::update_param ${f} subjects[0].namespace '{{ .Namespace }}'
				;;
			./ConfigMap_kubemacpool-mac-range-config.yaml)
				yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
				yaml-utils::update_param ${f} data.RANGE_START '{{ .RangeStart }}'
				yaml-utils::update_param ${f} data.RANGE_END '{{ .RangeEnd }}'
				;;
			./Service_kubemacpool-service.yaml)
				yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
				;;
			./Deployment_kubemacpool-cert-manager.yaml)
				yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
				yaml-utils::update_param ${f} spec.template.spec.containers[0].image '{{ .KubeMacPoolImage }}'
				yaml-utils::set_param ${f} spec.template.spec.containers[0].imagePullPolicy '{{ .ImagePullPolicy }}'
				yaml-utils::set_param ${f} spec.template.spec.nodeSelector '{{ toYaml .Placement.NodeSelector | nindent 8 }}'
				yaml-utils::set_param ${f} spec.template.spec.affinity '{{ toYaml .Placement.Affinity | nindent 8 }}'
				yaml-utils::set_param ${f} spec.template.spec.tolerations '{{ toYaml .Placement.Tolerations | nindent 8 }}'
				yaml-utils::set_param ${f} spec.template.spec.securityContext.runAsNonRoot '{{ .RunAsNonRoot }}'
				yaml-utils::set_param ${f} spec.template.spec.securityContext.runAsUser '{{ .RunAsUser }}'
				yaml-utils::set_param ${f} 'spec.template.metadata.annotations."openshift.io/required-scc"' 'restricted-v2'
				__set_empty_string_label ${f} 'spec.template.metadata.labels."allow-access-cluster-services"'
				__unquote_toYaml ${f}
				__unquote_numeric_and_boolean ${f}
				# Templatize cert rotation env var values with defaults
				sed -i '/- name: CA_ROTATE_INTERVAL/{n;s|value: .*|value: '\''{{ .CARotateInterval \| default "8760h0m0s" }}'\''|}' ${f}
				sed -i '/- name: CA_OVERLAP_INTERVAL/{n;s|value: .*|value: '\''{{ .CAOverlapInterval \| default "24h0m0s" }}'\''|}' ${f}
				sed -i '/- name: CERT_ROTATE_INTERVAL/{n;s|value: .*|value: '\''{{ .CertRotateInterval \| default "4380h0m0s" }}'\''|}' ${f}
				sed -i '/- name: CERT_OVERLAP_INTERVAL/{n;s|value: .*|value: '\''{{ .CertOverlapInterval \| default "24h0m0s" }}'\''|}' ${f}
				;;
			./Deployment_kubemacpool-mac-controller-manager.yaml)
				yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
				yaml-utils::update_param ${f} spec.template.spec.containers[0].image '{{ .KubeMacPoolImage }}'
				yaml-utils::set_param ${f} spec.template.spec.containers[0].imagePullPolicy '{{ .ImagePullPolicy }}'
				yaml-utils::set_param ${f} spec.template.spec.nodeSelector '{{ toYaml .Placement.NodeSelector | nindent 8 }}'
				yaml-utils::set_param ${f} spec.template.spec.affinity '{{ toYaml .Placement.Affinity | nindent 8 }}'
				yaml-utils::set_param ${f} spec.template.spec.tolerations '{{ toYaml .Placement.Tolerations | nindent 8 }}'
				yaml-utils::set_param ${f} spec.template.spec.securityContext.runAsNonRoot '{{ .RunAsNonRoot }}'
				yaml-utils::set_param ${f} spec.template.spec.securityContext.runAsUser '{{ .RunAsUser }}'
				yaml-utils::set_param ${f} 'spec.template.metadata.annotations."openshift.io/required-scc"' 'restricted-v2'
				__set_empty_string_label ${f} 'spec.template.metadata.labels."allow-access-cluster-services"'
				__unquote_toYaml ${f}
				__unquote_numeric_and_boolean ${f}
				# Templatize TLS args
				sed -i '/            - --wait-time=/a\            - "--tls-min-version={{ .TLSMinVersion }}"' ${f}
				sed -i '/            - "--tls-min-version={{ .TLSMinVersion }}"/a{{ if index . "TLSSecurityProfileCiphers" }}\n            - "--tls-cipher-suites={{ .TLSSecurityProfileCiphers }}"\n{{ end }}' ${f}
				;;
			./NetworkPolicy_kubemacpool-allow-ingress-to-metrics-endpoint.yaml | \
			./NetworkPolicy_kubemacpool-allow-ingress-to-webhook.yaml)
				yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
				;;
			./MutatingWebhookConfiguration_kubemacpool-mutator.yaml)
				sed -i "s/namespace: kubemacpool-system/namespace: '{{ .Namespace }}'/g" ${f}
				;;
			./Role_kubemacpool-prometheus.yaml)
				yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
				;;
			./RoleBinding_kubemacpool-prometheus.yaml)
				yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
				yaml-utils::update_param ${f} subjects[0].name '{{ .MonitoringServiceAccount }}'
				yaml-utils::update_param ${f} subjects[0].namespace '{{ .MonitoringNamespace }}'
				;;
			./Service_kubemacpool-metrics-service.yaml)
				yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
				;;
			./PrometheusRule_kubemacpool-prometheus-rule.yaml)
				yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
				# Escape Go template expressions ({{ $value }} -> {{ "{{" }} $value {{ "}}" }})
				sed -i 's/{{ \$value }}/{{ "{{" }} $value {{ "}}" }}/g' ${f}
				;;
			./ServiceMonitor_kubemacpool-metrics-monitor.yaml)
				yaml-utils::update_param ${f} metadata.namespace '{{ .Namespace }}'
				;;
		esac
	done
}

echo 'Bumping kubemacpool'
KUBEMACPOOL_URL=$(yaml-utils::get_component_url kubemacpool)
KUBEMACPOOL_COMMIT=$(yaml-utils::get_component_commit kubemacpool)
KUBEMACPOOL_REPO=$(yaml-utils::get_component_repo ${KUBEMACPOOL_URL})

TEMP_DIR=$(git-utils::create_temp_path kubemacpool)
trap "rm -rf ${TEMP_DIR}" EXIT
KUBEMACPOOL_PATH=${TEMP_DIR}/${KUBEMACPOOL_REPO}

echo 'Fetch kubemacpool sources'
git-utils::fetch_component ${KUBEMACPOOL_PATH} ${KUBEMACPOOL_URL} ${KUBEMACPOOL_COMMIT}

echo 'Adjust kubemacpool manifests for CNAO'
(
	cd ${KUBEMACPOOL_PATH}
	mkdir -p config/cnao
	cp config/release/kubemacpool.yaml config/cnao/

	echo 'Split manifest per object'
	cd config/cnao
	$(yaml-utils::append_delimiter kubemacpool.yaml)
	$(yaml-utils::split_yaml_by_seperator . kubemacpool.yaml)
	rm kubemacpool.yaml
	$(yaml-utils::rename_files_by_object .)

	echo 'Parametize manifests by object'
	__parametize_by_object

	echo 'Rejoin sub-manifests to final manifest'
	cat Namespace_kubemacpool-system.yaml \
		ServiceAccount_kubemacpool-sa.yaml \
		ClusterRole_kubemacpool-manager-role.yaml \
		ClusterRoleBinding_kubemacpool-manager-rolebinding.yaml \
		ConfigMap_kubemacpool-mac-range-config.yaml \
		Service_kubemacpool-service.yaml \
		Deployment_kubemacpool-cert-manager.yaml \
		Deployment_kubemacpool-mac-controller-manager.yaml \
		NetworkPolicy_kubemacpool-allow-ingress-to-metrics-endpoint.yaml \
		NetworkPolicy_kubemacpool-allow-ingress-to-webhook.yaml \
		MutatingWebhookConfiguration_kubemacpool-mutator.yaml \
		> kubemacpool.yaml
)

rm -rf data/kubemacpool/*
cp ${KUBEMACPOOL_PATH}/config/cnao/kubemacpool.yaml data/kubemacpool/

echo 'Prepare kubemacpool monitoring manifest'
(
	cd ${KUBEMACPOOL_PATH}
	mkdir -p config/cnao-monitoring
	cp config/release/kubemacpool-monitoring.yaml config/cnao-monitoring/

	echo 'Split monitoring manifest per object'
	cd config/cnao-monitoring
	$(yaml-utils::append_delimiter kubemacpool-monitoring.yaml)
	$(yaml-utils::split_yaml_by_seperator . kubemacpool-monitoring.yaml)
	rm kubemacpool-monitoring.yaml
	$(yaml-utils::rename_files_by_object .)

	echo 'Parametize monitoring manifests by object'
	__parametize_by_object

	echo 'Rejoin monitoring sub-manifests to final manifest'
	echo '{{ if .MonitoringAvailable }}' > kubemacpool-monitoring.yaml
	cat Role_kubemacpool-prometheus.yaml \
		RoleBinding_kubemacpool-prometheus.yaml \
		Service_kubemacpool-metrics-service.yaml \
		PrometheusRule_kubemacpool-prometheus-rule.yaml \
		ServiceMonitor_kubemacpool-metrics-monitor.yaml \
		>> kubemacpool-monitoring.yaml
	echo '{{ end }}' >> kubemacpool-monitoring.yaml
)

cp ${KUBEMACPOOL_PATH}/config/cnao-monitoring/kubemacpool-monitoring.yaml data/kubemacpool/

echo 'Get kubemacpool image name and update it under CNAO'
KUBEMACPOOL_TAG=$(git-utils::get_component_tag ${KUBEMACPOOL_PATH})
KUBEMACPOOL_IMAGE=quay.io/kubevirt/kubemacpool
KUBEMACPOOL_IMAGE_TAGGED=${KUBEMACPOOL_IMAGE}:${KUBEMACPOOL_TAG}
KUBEMACPOOL_IMAGE_DIGEST="$(docker-utils::get_image_digest "${KUBEMACPOOL_IMAGE_TAGGED}" "${KUBEMACPOOL_IMAGE}")"

sed -i -r "s#\"${KUBEMACPOOL_IMAGE}(@sha256)?:.*\"#\"${KUBEMACPOOL_IMAGE_DIGEST}\"#" pkg/components/components.go
sed -i -r "s#\"${KUBEMACPOOL_IMAGE}(@sha256)?:.*\"#\"${KUBEMACPOOL_IMAGE_DIGEST}\"#" test/releases/${CNAO_VERSION}.go
