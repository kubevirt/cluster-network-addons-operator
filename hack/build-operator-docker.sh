#!/bin/bash

if [ -z "$ARCH" ] || [ -z "$PLATFORMS" ] || [ -z "$OPERATOR_IMAGE_TAGGED" ]; then
    echo "Error: ARCH, PLATFORMS and OPERATOR_IMAGE_TAGGED must be set."
    exit 1
fi

IFS=',' read -r -a PLATFORM_LIST <<< "$PLATFORMS"

GO_VERSION=$(./hack/go-version.sh | cut -d'.' -f1-2)

BUILD_ARGS="--build-arg BUILD_ARCH=$ARCH --build-arg GO_VERSION=$GO_VERSION -f build/operator/Dockerfile -t $OPERATOR_IMAGE_TAGGED --push ."

if [ ${#PLATFORM_LIST[@]} -eq 1 ]; then
    docker build --platform "$PLATFORMS" $BUILD_ARGS
else
    ./hack/init-buildx.sh "$DOCKER_BUILDER"
    docker buildx build --platform "$PLATFORMS" $BUILD_ARGS
    docker buildx rm "$DOCKER_BUILDER" 2>/dev/null || echo "Builder ${DOCKER_BUILDER} not found or already removed, skipping."
fi
