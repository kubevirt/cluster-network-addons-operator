#!/usr/bin/env bash

set -xe

source hack/components/yaml-utils.sh
source hack/components/git-utils.sh
source hack/components/docker-utils.sh

export OCI_BIN=${OCI_BIN:-$(docker-utils::determine_cri_bin)}
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
DOCKER_BUILDER="${DOCKER_BUILDER:-linux-bridge-docker-builder}"
# By default, the build will target all supported platforms(i.e PLATFORM_LIST).
# To build for specific platforms, you can:
# 1. Specify individual platforms:
#    export PLATFORMS=linux/amd64
#    or
#    export PLATFORMS=linux/amd64,linux/arm64
PLATFORM_LIST="linux/amd64,linux/s390x,linux/arm64"
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
PLATFORMS="${PLATFORMS:-all}"
[ "$PLATFORMS" == "all" ] && PLATFORMS="${PLATFORM_LIST}"
IFS=',' read -r -a PLATFORM_LIST <<< "$PLATFORMS"

create_dockerfile() {
    cat <<EOF > Dockerfile
FROM --platform=\$BUILDPLATFORM registry.access.redhat.com/ubi8/ubi-minimal AS builder
ARG TARGETOS
ARG TARGETARCH
RUN microdnf install -y golang git
RUN \
    git clone https://${LINUX_BRIDGE_REPO} ${LINUX_BRIDGE_PATH} && \
    cd ${LINUX_BRIDGE_PATH} && \
    git checkout ${LINUX_BRIDGE_TAG}
WORKDIR ${LINUX_BRIDGE_PATH}
RUN GOFLAGS=-mod=vendor GOARCH=\${TARGETARCH} GOOS=\${TARGETOS} ./build_linux.sh

FROM --platform=linux/\$TARGETARCH registry.access.redhat.com/ubi8/ubi-minimal
LABEL org.opencontainers.image.authors="phoracek@redhat.com"
ENV SOURCE_DIR=${REMOTE_SOURCE_DIR}/app
RUN mkdir -p ${LINUX_BRIDGE_TAR_CONTAINER_DIR}
RUN microdnf install -y findutils
COPY --from=builder ${LINUX_BRIDGE_PATH}/bin/bridge ${LINUX_BRIDGE_TAR_CONTAINER_DIR}/bridge
COPY --from=builder ${LINUX_BRIDGE_PATH}/bin/tuning ${LINUX_BRIDGE_TAR_CONTAINER_DIR}/tuning
RUN sha256sum ${LINUX_BRIDGE_TAR_CONTAINER_DIR}/bridge >${LINUX_BRIDGE_TAR_CONTAINER_DIR}/bridge.checksum
RUN sha256sum ${LINUX_BRIDGE_TAR_CONTAINER_DIR}/tuning >${LINUX_BRIDGE_TAR_CONTAINER_DIR}/tuning.checksum
EOF
}

# Retry a command with exponential backoff
# Usage: retry_with_backoff <max_attempts> <initial_delay> <command> [args...]
retry_with_backoff() {
    local max_attempts=$1
    local delay=$2
    shift 2
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        echo "Attempt $attempt of $max_attempts: $*"
        if "$@"; then
            echo "Command succeeded on attempt $attempt"
            return 0
        fi

        if [ $attempt -lt $max_attempts ]; then
            echo "Command failed. Waiting ${delay}s before retry..."
            sleep "$delay"
            delay=$((delay * 2))
        fi
        attempt=$((attempt + 1))
    done

    echo "Command failed after $max_attempts attempts"
    return 1
}

check_and_create_docker_builder() {
    existing_builder=$(docker buildx ls | grep -w "$DOCKER_BUILDER" | awk '{print $1}' || true)
    if [ -n "$existing_builder" ]; then
        echo "Builder '$DOCKER_BUILDER' already exists. Using existing builder."
        docker buildx use "$DOCKER_BUILDER"
    else
        echo "Creating a new Docker Buildx builder: $DOCKER_BUILDER"
        docker buildx create --driver-opt network=host --use --name "$DOCKER_BUILDER"
    fi
}

build_docker_image() {
    retry_with_backoff 3 5 docker buildx build --platform "${PLATFORMS}" -t "${LINUX_BRIDGE_IMAGE_TAGGED}" . --push
    docker buildx rm "$DOCKER_BUILDER"
}

build_podman_image() {
    podman manifest rm "${LINUX_BRIDGE_IMAGE_TAGGED}" 2>/dev/null || true
    podman rmi "${LINUX_BRIDGE_IMAGE_TAGGED}" 2>/dev/null || true
    podman rmi $(podman images --filter "dangling=true" -q) 2>/dev/null || true
    podman manifest create "${LINUX_BRIDGE_IMAGE_TAGGED}"

    for platform in "${PLATFORM_LIST[@]}"; do
        retry_with_backoff 3 5 podman build --platform "$platform" --manifest "${LINUX_BRIDGE_IMAGE_TAGGED}" .
    done
}

push_image_to_kubevirt_repo() {
    echo 'Push the image to KubeVirt repo'
    if [ "${OCI_BIN}" == "podman" ]; then
        if [ ! -z "${PUSH_IMAGES}" ]; then
            retry_with_backoff 3 5 podman manifest push "${LINUX_BRIDGE_IMAGE_TAGGED}"
        fi
    fi
}


(
    cd ${LINUX_BRIDGE_PATH}
    create_dockerfile
    (
        if [[ "${OCI_BIN}" == "docker" ]]; then
            check_and_create_docker_builder
            build_docker_image
        elif [[ "${OCI_BIN}" == "podman" ]]; then
            build_podman_image
            push_image_to_kubevirt_repo
        else
            echo "Invalid OCI_BIN value. It must be either 'docker' or 'podman'."
            exit 1
        fi
    )
)

if [[ -n "$(docker-utils::check_image_exists "${LINUX_BRIDGE_IMAGE}" "${LINUX_BRIDGE_TAG}")" ]]; then
    LINUX_BRIDGE_IMAGE_DIGEST="$(docker-utils::get_image_digest "${LINUX_BRIDGE_IMAGE_TAGGED}" "${LINUX_BRIDGE_IMAGE}")"
else
    LINUX_BRIDGE_IMAGE_DIGEST=${LINUX_BRIDGE_IMAGE_TAGGED}
fi

echo 'Update linux-bridge references under CNAO'
sed -i -r "s#\"${LINUX_BRIDGE_IMAGE}(@sha256)?:.*\"#\"${LINUX_BRIDGE_IMAGE_DIGEST}\"#" pkg/components/components.go
sed -i -r "s#\"${LINUX_BRIDGE_IMAGE}(@sha256)?:.*\"#\"${LINUX_BRIDGE_IMAGE_DIGEST}\"#" test/releases/${CNAO_VERSION}.go
