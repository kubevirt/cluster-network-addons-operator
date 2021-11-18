#!/usr/bin/env bash

set -xeo pipefail

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh
source hack/components/podman-utils.sh

echo 'Bumping kubernetes-nmstate'
NMSTATE_URL=$(yaml-utils::get_component_url nmstate)
NMSTATE_COMMIT=$(yaml-utils::get_component_commit nmstate)
NMSTATE_REPO=$(yaml-utils::get_component_repo ${NMSTATE_URL})

TEMP_DIR=$(git-utils::create_temp_path nmstate)
trap "rm -rf ${TEMP_DIR}" EXIT
NMSTATE_PATH=${TEMP_DIR}/${NMSTATE_REPO}

echo 'Fetch kubernetes-nmstate sources'
git-utils::fetch_component ${NMSTATE_PATH} ${NMSTATE_URL} ${NMSTATE_COMMIT}

echo 'Get kubernetes-nmstate image name and update it under CNAO'
NMSTATE_TAG=$(git-utils::get_component_tag ${NMSTATE_PATH})
NMSTATE_IMAGE=quay.io/nmstate/kubernetes-nmstate-handler
NMSTATE_IMAGE_TAGGED=${NMSTATE_IMAGE}:${NMSTATE_TAG}
NMSTATE_IMAGE_DIGEST="$(podman-utils::get_image_digest "${NMSTATE_IMAGE_TAGGED}" "${NMSTATE_IMAGE}")"

sed -i -r "s#\"${NMSTATE_IMAGE}(@sha256)?:.*\"#\"${NMSTATE_IMAGE_DIGEST}\"#" pkg/components/components.go
sed -i -r "s#\"${NMSTATE_IMAGE}(@sha256)?:.*\"#\"${NMSTATE_IMAGE_DIGEST}\"#" "test/releases/${CNAO_VERSION}.go"

echo 'Configure nmstate-webhook and nmstate-handler templates and save the rendered manifest under CNAO data'
(
    cd ${NMSTATE_PATH}
    mkdir -p config/cnao/handler

    cp $NMSTATE_PATH/deploy/handler/* config/cnao/handler

    sed -i \
        -e "s#WebhookAffinity#PlacementConfiguration.Infra.Affinity#" \
        -e "s#InfraTolerations#PlacementConfiguration.Infra.Tolerations#" \
        -e "s#InfraNodeSelector#PlacementConfiguration.Infra.NodeSelector#" \
        -e "s#HandlerAffinity#PlacementConfiguration.Workloads.Affinity#" \
        -e "s#HandlerTolerations#PlacementConfiguration.Workloads.Tolerations#" \
        -e "s#HandlerNodeSelector#PlacementConfiguration.Workloads.NodeSelector#" \
        config/cnao/handler/operator.yaml
)

echo 'Copy kubernetes-nmstate manifests'
rm -rf data/nmstate/*
mkdir -p data/nmstate/{operator,operand}/
cp $NMSTATE_PATH/config/cnao/handler/* data/nmstate/operand/
cp $NMSTATE_PATH/deploy/crds/nmstate.io_nodenetwork*.yaml data/nmstate/operand/
cp $NMSTATE_PATH/deploy/openshift/scc.yaml data/nmstate/operand/scc.yaml
sed -i "s/---/{{ if .EnableSCC }}\n---/" data/nmstate/operand/scc.yaml
echo "{{ end }}" >> data/nmstate/operand/scc.yaml

cp $NMSTATE_PATH/deploy/crds/nmstate.io_*_nmstate_cr.yaml data/nmstate/operator/

echo 'Apply custom CNAO patches on kubernetes-nmstate manifests'
sed -i -z 's#kind: Secret\nmetadata:#kind: Secret\nmetadata:\n  annotations:\n    networkaddonsoperator.network.kubevirt.io\/rejectOwner: ""#' data/nmstate/operand/operator.yaml
