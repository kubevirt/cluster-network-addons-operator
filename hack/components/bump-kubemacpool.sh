#!/usr/bin/env bash

set -xeo pipefail

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh

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
- cnao_image_patch.yaml
- cnao_rejectowner_patch.yaml
EOF

    cat <<EOF > config/cnao/cnao_image_patch.yaml
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
          - name: CA_ROTATE_INTERVAL
            value: "{{ .CARotateInterval }}"
          - name: CA_OVERLAP_INTERVAL
            value: "{{ .CAOverlapInterval }}"
          - name: CERT_ROTATE_INTERVAL
            value: "{{ .CertRotateInterval }}"
EOF

    cat <<EOF > config/cnao/cnao_rejectowner_patch.yaml
apiVersion: v1
kind: Secret
metadata:
  name: service
  namespace: system
  annotations:
    networkaddonsoperator.network.kubevirt.io/rejectOwner: ""
EOF
)

rm -rf data/kubemacpool/*
(
    cd $KUBEMACPOOL_PATH
    make tools 1>/dev/null
    ./build/_output/bin/go/bin/kustomize build config/cnao | \
        sed 's/kubemacpool-system/{{ .Namespace }}/' | \
        sed 's/RANGE_START: .*/RANGE_START: {{ .RangeStart }}/' | \
        sed 's/RANGE_END: .*/RANGE_END: {{ .RangeEnd }}/'
) > data/kubemacpool/kubemacpool.yaml

echo 'Get kubemacpool image name and update it under CNAO'
KUBEMACPOOL_TAG=$(git-utils::get_component_tag ${KUBEMACPOOL_PATH})
KUBEMACPOOL_IMAGE=quay.io/kubevirt/kubemacpool
KUBEMACPOOL_IMAGE_TAGGED=${KUBEMACPOOL_IMAGE}:${KUBEMACPOOL_TAG}
sed -i "s#\"${KUBEMACPOOL_IMAGE}:.*\"#\"${KUBEMACPOOL_IMAGE_TAGGED}\"#" pkg/components/components.go
sed -i "s#\"${KUBEMACPOOL_IMAGE}:.*\"#\"${KUBEMACPOOL_IMAGE_TAGGED}\"#" test/releases/${CNAO_VERSION}.go
