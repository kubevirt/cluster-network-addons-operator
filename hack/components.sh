#!/usr/bin/env bash

set -xe

CNAO_VERSION=${VERSION} # Exported from Makefile

echo 'Setup temporary Go path'
export GOPATH=${PWD}/_components/go
mkdir -p $GOPATH
export PATH=${GOPATH}/bin:${PATH}

echo 'kubemacpool'
KUBEMACPOOL_URL=$(cat components.yaml | shyaml get-value components.kubemacpool.url)
KUBEMACPOOL_COMMIT=$(cat components.yaml | shyaml get-value components.kubemacpool.commit)
KUBEMACPOOL_REPO=$(echo ${KUBEMACPOOL_URL} | sed 's#https://\(.*\)#\1#')
KUBEMACPOOL_PATH=${GOPATH}/src/${KUBEMACPOOL_REPO}

echo 'Fetch kubemacpool sources'
(
    go get ${KUBEMACPOOL_REPO} || true
    cd ${KUBEMACPOOL_PATH}
    git fetch origin
    git reset --hard
    git checkout ${KUBEMACPOOL_COMMIT}
)

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
EOF

    cat <<EOF > config/cnao/cnao_image_patch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mac-controller-manager
spec:
  template:
    spec:
      containers:
      - image: "{{ .KubeMacPoolImage }}"
        imagePullPolicy: "{{ .ImagePullPolicy }}"
        name: manager
EOF
)
rm -rf data/kubemacpool/*
(
    cd $KUBEMACPOOL_PATH
    kustomize build config/cnao | \
        sed 's/name: kubemacpool-system/name: {{ .Namespace }}/' | \
        sed 's/RANGE_START: .*/RANGE_START: {{ .RangeStart }}/' | \
        sed 's/RANGE_END: .*/RANGE_END: {{ .RangeEnd }}/'
) > data/kubemacpool/kubemacpool.yaml

echo 'Get kubemacpool image name and update it under CNAO'
KUBEMACPOOL_TAG=$(
    cd $KUBEMACPOOL_PATH
    git describe --tags
)
KUBEMACPOOL_IMAGE=quay.io/kubevirt/kubemacpool
KUBEMACPOOL_IMAGE_TAGGED=${KUBEMACPOOL_IMAGE}:${KUBEMACPOOL_TAG}
sed -i "s#\"${KUBEMACPOOL_IMAGE}:.*\"#\"${KUBEMACPOOL_IMAGE_TAGGED}\"#" pkg/components/components.go
sed -i "s#\"${KUBEMACPOOL_IMAGE}:.*\"#\"${KUBEMACPOOL_IMAGE_TAGGED}\"#" test/releases/${CNAO_VERSION}.go
