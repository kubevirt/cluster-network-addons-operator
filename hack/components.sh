#!/usr/bin/env bash

set -xe

function fetch_component() {
    destination=$1
    url=$2
    commit=$3

    if [ ! -d ${destination} ]; then
        mkdir -p ${destination}
        git clone ${url} ${destination}
    fi

    (
        cd ${destination}
        git fetch origin
        git reset --hard
        git checkout ${commit}
    )
}

function get_component_tag() {
    component_dir=$1

    (
        cd ${component_dir}
        git describe --tags
    )
}

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
fetch_component ${KUBEMACPOOL_PATH} ${KUBEMACPOOL_URL} ${KUBEMACPOOL_COMMIT}

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
  namespace: system
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
        sed 's/kubemacpool-system/{{ .Namespace }}/' | \
        sed 's/RANGE_START: .*/RANGE_START: {{ .RangeStart }}/' | \
        sed 's/RANGE_END: .*/RANGE_END: {{ .RangeEnd }}/'
) > data/kubemacpool/kubemacpool.yaml

echo 'Get kubemacpool image name and update it under CNAO'
KUBEMACPOOL_TAG=$(get_component_tag ${KUBEMACPOOL_PATH})
KUBEMACPOOL_IMAGE=quay.io/kubevirt/kubemacpool
KUBEMACPOOL_IMAGE_TAGGED=${KUBEMACPOOL_IMAGE}:${KUBEMACPOOL_TAG}
sed -i "s#\"${KUBEMACPOOL_IMAGE}:.*\"#\"${KUBEMACPOOL_IMAGE_TAGGED}\"#" pkg/components/components.go
sed -i "s#\"${KUBEMACPOOL_IMAGE}:.*\"#\"${KUBEMACPOOL_IMAGE_TAGGED}\"#" test/releases/${CNAO_VERSION}.go

echo 'macvtap'
MACVTAP_URL=$(cat components.yaml | shyaml get-value components.macvtap-cni.url)
MACVTAP_COMMIT=$(cat components.yaml | shyaml get-value components.macvtap-cni.commit)
MACVTAP_REPO=$(echo ${MACVTAP_URL} | sed 's#https://\(.*\)#\1#')
MACVTAP_PATH=${GOPATH}/src/${MACVTAP_REPO}

echo 'Fetch macvtap-cni sources'
fetch_component ${MACVTAP_PATH} ${MACVTAP_URL} ${MACVTAP_COMMIT}

rm -rf data/macvtap/*
echo 'Copy the templates from the macvtap-cni repo ...'
cp ${MACVTAP_PATH}/templates/namespace.yaml.in data/macvtap/000-ns.yaml
echo "{{ if .EnableSCC }}" >> data/macvtap/001-rbac.yaml
cat ${MACVTAP_PATH}/templates/scc.yaml.in >> data/macvtap/001-rbac.yaml
echo "{{ end }}" >> data/macvtap/001-rbac.yaml
cat <<EOF > data/macvtap/002-macvtap-daemonset.yaml
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: macvtap-deviceplugin-config
  namespace: {{ .Namespace }}
data:
  DP_MACVTAP_CONF: "[]"
---
EOF
cat ${MACVTAP_PATH}/templates/macvtap.yaml.in >> data/macvtap/002-macvtap-daemonset.yaml

echo 'Get macvtap-cni image name and update it under CNAO'
MACVTAP_TAG=$(get_component_tag ${MACVTAP_PATH})
MACVTAP_IMAGE=quay.io/kubevirt/macvtap-cni
MACVTAP_IMAGE_TAGGED=${MACVTAP_IMAGE}:${MACVTAP_TAG}
sed -i "s#\"${MACVTAP_IMAGE}:.*\"#\"${MACVTAP_IMAGE_TAGGED}\"#" pkg/components/components.go
# TODO: uncomment the following line *once* there is macvtap upgrade is supported
#sed -i "s#\"${MACVTAP_IMAGE}:.*\"#\"${MACVTAP_IMAGE_TAGGED}\"#" test/releases/${CNAO_VERSION}.go

echo 'Clone kubernetes-nmstate'
NMSTATE_URL=$(cat components.yaml | shyaml get-value components.nmstate.url)
NMSTATE_COMMIT=$(cat components.yaml | shyaml get-value components.nmstate.commit)
NMSTATE_REPO=$(echo ${NMSTATE_URL} | sed 's#https://\(.*\)#\1#')
NMSTATE_PATH=${GOPATH}/src/${NMSTATE_REPO}
fetch_component ${NMSTATE_PATH} ${NMSTATE_URL} ${NMSTATE_COMMIT}

echo 'Get kubernetes-nmstate image name and update it under CNAO'
NMSTATE_TAG=$(get_component_tag ${NMSTATE_PATH})
NMSTATE_IMAGE=quay.io/nmstate/kubernetes-nmstate-handler
NMSTATE_IMAGE_TAGGED=${NMSTATE_IMAGE}:${NMSTATE_TAG}
sed -i "s#\"${NMSTATE_IMAGE}:.*\"#\"${NMSTATE_IMAGE_TAGGED}\"#" \
    pkg/components/components.go \
    test/releases/${CNAO_VERSION}.go

echo 'Copy kubernetes-nmstate manifests'
rm -rf data/nmstate/*
cp $NMSTATE_PATH/deploy/handler/* data/nmstate/
cp $NMSTATE_PATH/deploy/crds/*nodenetwork*crd* data/nmstate/
