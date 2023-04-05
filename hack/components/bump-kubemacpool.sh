#!/usr/bin/env bash

set -xeo pipefail

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh
source hack/components/docker-utils.sh

echo 'Bumping kubemacpool'
KUBEMACPOOL_URL=$(yaml-utils::get_component_url kubemacpool)
KUBEMACPOOL_COMMIT=$(yaml-utils::get_component_commit kubemacpool)
KUBEMACPOOL_REPO=$(yaml-utils::get_component_repo ${KUBEMACPOOL_URL})

TEMP_DIR=$(git-utils::create_temp_path kubemacpool)
trap "rm -rf ${TEMP_DIR}" EXIT
KUBEMACPOOL_PATH=${TEMP_DIR}/${KUBEMACPOOL_REPO}

echo 'Fetch kubemacpool sources'
git-utils::fetch_component ${KUBEMACPOOL_PATH} ${KUBEMACPOOL_URL} ${KUBEMACPOOL_COMMIT}

echo 'Configure kustomize for CNAO templates and save the rendered manifest under CNAO data'
(
    cd ${KUBEMACPOOL_PATH}
    mkdir -p config/cnao

    cat <<EOF > config/cnao/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: "{{ .Namespace }}"
bases:
- ../default
patchesStrategicMerge:
- cnao_kubemacpool_manager_patch.yaml
- cnao_cert-manager_patch.yaml
- cnao_placement_patch.yaml
- mutatevirtualmachines_opt_mode_patch.yaml
- mutatepods_opt_mode_patch.yaml
EOF

    cat <<EOF > config/cnao/cnao_kubemacpool_manager_patch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mac-controller-manager
  namespace: system
spec:
  template:
    spec:
      containers:
      - image: "{{ .KubeMacPoolImage }}"
        imagePullPolicy: "{{ .ImagePullPolicy }}"
        name: manager
        env:
          - name: TLS_MIN_VERSION
            value: "{{ .TLSMinVersion }}"
          - name: TLS_CIPHERS
            value: "{{ .TLSSecurityProfileCiphers }}"
      - image: "{{ .KubeRbacProxyImage }}"
        imagePullPolicy: "{{ .ImagePullPolicy }}"
        name: kube-rbac-proxy
EOF

    cat <<EOF > config/cnao/cnao_cert-manager_patch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cert-manager
  namespace: system
spec:
  template:
    spec:
      affinity: AFFINITY
      nodeSelector: NODE_SELECTOR
      tolerations: TOLERATIONS
      containers:
      - image: "{{ .KubeMacPoolImage }}"
        imagePullPolicy: "{{ .ImagePullPolicy }}"
        name: manager
        env:
          - name: CA_ROTATE_INTERVAL
            value: "{{ .CARotateInterval | default \"8760h0m0s\" }}"
          - name: CA_OVERLAP_INTERVAL
            value: "{{ .CAOverlapInterval | default \"24h0m0s\" }}"
          - name: CERT_ROTATE_INTERVAL
            value: "{{ .CertRotateInterval | default \"4380h0m0s\" }}"
          - name: CERT_OVERLAP_INTERVAL
            value: "{{ .CertOverlapInterval | default \"24h0m0s\" }}"
EOF

    cat <<EOF > config/cnao/cnao_placement_patch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mac-controller-manager
  namespace: system
spec:
  template:
    spec:
      affinity: AFFINITY
      nodeSelector: NODE_SELECTOR
      tolerations: TOLERATIONS
EOF

    (
        cd config/cnao

        echo setting pods to opt-in mode
        cp ../default/mutatepods_opt_in_patch.yaml mutatepods_opt_mode_patch.yaml
        echo setting vms to opt-in mode
        cp ../default/mutatevirtualmachines_opt_out_patch.yaml mutatevirtualmachines_opt_mode_patch.yaml
    )

)

rm -rf data/kubemacpool/*
(
    cd $KUBEMACPOOL_PATH
    make tools 1>/dev/null
    ./build/_output/bin/go/bin/kustomize build config/cnao | \
        sed 's/kubemacpool-system/{{ .Namespace }}/' | \
        sed 's/RANGE_START: .*/RANGE_START: {{ .RangeStart }}/' | \
        sed 's/RANGE_END: .*/RANGE_END: {{ .RangeEnd }}/' | \
        sed 's/AFFINITY/{{ toYaml .Placement.Affinity | nindent 8 }}/' | \
        sed 's/NODE_SELECTOR/{{ toYaml .Placement.NodeSelector | nindent 8 }}/' | \
        sed 's/TOLERATIONS/{{ toYaml .Placement.Tolerations | nindent 8 }}/'
) > data/kubemacpool/kubemacpool.yaml

echo 'Get kubemacpool image name and update it under CNAO'
KUBEMACPOOL_TAG=$(git-utils::get_component_tag ${KUBEMACPOOL_PATH})
KUBEMACPOOL_IMAGE=quay.io/kubevirt/kubemacpool
KUBEMACPOOL_IMAGE_TAGGED=${KUBEMACPOOL_IMAGE}:${KUBEMACPOOL_TAG}
KUBEMACPOOL_IMAGE_DIGEST="$(docker-utils::get_image_digest "${KUBEMACPOOL_IMAGE_TAGGED}" "${KUBEMACPOOL_IMAGE}")"

sed -i -r "s#\"${KUBEMACPOOL_IMAGE}(@sha256)?:.*\"#\"${KUBEMACPOOL_IMAGE_DIGEST}\"#" pkg/components/components.go
sed -i -r "s#\"${KUBEMACPOOL_IMAGE}(@sha256)?:.*\"#\"${KUBEMACPOOL_IMAGE_DIGEST}\"#" test/releases/${CNAO_VERSION}.go
