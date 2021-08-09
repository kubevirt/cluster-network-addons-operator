#!/usr/bin/env bash

set -xe

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh
source hack/components/docker-utils.sh

echo Bumping linux-bridge
LINUX_BRIDGE_URL=$(yaml-utils::get_component_url linux-bridge)
LINUX_BRIDGE_COMMIT=$(yaml-utils::get_component_commit linux-bridge)
LINUX_BRIDGE_REPO=$(yaml-utils::get_component_repo ${LINUX_BRIDGE_URL})

TEMP_DIR=$(git-utils::create_temp_path linux-bridge)
trap "rm -rf ${TEMP_DIR}" EXIT
LINUX_BRIDGE_PATH=${TEMP_DIR}/${LINUX_BRIDGE_REPO}

echo 'Fetch linux-bridge sources'
git-utils::fetch_component ${LINUX_BRIDGE_PATH} ${LINUX_BRIDGE_URL} ${LINUX_BRIDGE_COMMIT}

echo 'Get linux-bridge tag'
LINUX_BRIDGE_TAG=$(git-utils::get_component_tag ${LINUX_BRIDGE_PATH})

echo 'Build container image with linux-bridge binaries'
LINUX_BRIDGE_TAR_CONTAINER_DIR=/usr/src/github.com/containernetworking/plugins/bin
LINUX_BRIDGE_IMAGE=quay.io/kubevirt/cni-default-plugins
LINUX_BRIDGE_IMAGE_TAGGED=${LINUX_BRIDGE_IMAGE}:${LINUX_BRIDGE_TAG}
(
    cd ${LINUX_BRIDGE_PATH}
    cat <<EOF > Dockerfile
FROM registry.access.redhat.com/ubi8/ubi-minimal AS builder
RUN microdnf install -y golang git
RUN \
    git clone https://${LINUX_BRIDGE_REPO} ${LINUX_BRIDGE_PATH} && \
    cd ${LINUX_BRIDGE_PATH} && \
    git checkout ${LINUX_BRIDGE_TAG}
WORKDIR ${LINUX_BRIDGE_PATH}
RUN GOFLAGS=-mod=vendor ./build_linux.sh

FROM registry.access.redhat.com/ubi8/ubi-minimal
ENV SOURCE_DIR=${REMOTE_SOURCE_DIR}/app
RUN mkdir -p ${LINUX_BRIDGE_TAR_CONTAINER_DIR}
RUN microdnf install -y findutils
COPY --from=builder ${LINUX_BRIDGE_PATH}/bin/bridge ${LINUX_BRIDGE_TAR_CONTAINER_DIR}/bridge
COPY --from=builder ${LINUX_BRIDGE_PATH}/bin/tuning ${LINUX_BRIDGE_TAR_CONTAINER_DIR}/tuning
RUN sha256sum ${LINUX_BRIDGE_TAR_CONTAINER_DIR}/bridge >${LINUX_BRIDGE_TAR_CONTAINER_DIR}/bridge.checksum
RUN sha256sum ${LINUX_BRIDGE_TAR_CONTAINER_DIR}/tuning >${LINUX_BRIDGE_TAR_CONTAINER_DIR}/tuning.checksum
EOF
    docker build -t ${LINUX_BRIDGE_IMAGE_TAGGED} .
)

echo 'Push the image to KubeVirt repo'
(
    if [ ! -z ${PUSH_IMAGES} ]; then
        docker push "${LINUX_BRIDGE_IMAGE_TAGGED}"
    fi
)

if [[ -n "$(docker-utils::check_image_exists "${LINUX_BRIDGE_IMAGE}" "${LINUX_BRIDGE_TAG}")" ]]; then
    LINUX_BRIDGE_IMAGE_DIGEST="$(docker-utils::get_image_digest "${LINUX_BRIDGE_IMAGE_TAGGED}" "${LINUX_BRIDGE_IMAGE}")"
else
    LINUX_BRIDGE_IMAGE_DIGEST=${LINUX_BRIDGE_IMAGE_TAGGED}
fi

echo 'Update linux-bridge references under CNAO'
sed -i -r "s#\"${LINUX_BRIDGE_IMAGE}(@sha256)?:.*\"#\"${LINUX_BRIDGE_IMAGE_DIGEST}\"#" pkg/components/components.go
sed -i -r "s#\"${LINUX_BRIDGE_IMAGE}(@sha256)?:.*\"#\"${LINUX_BRIDGE_IMAGE_DIGEST}\"#" test/releases/${CNAO_VERSION}.go
