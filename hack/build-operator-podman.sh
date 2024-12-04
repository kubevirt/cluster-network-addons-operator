#!/bin/bash

if [ -z "$ARCH" ] || [ -z "$PLATFORMS" ] || [ -z "$OPERATOR_IMAGE_TAGGED" ]; then
    echo "Error: ARCH, PLATFORMS, and OPERATOR_IMAGE_TAGGED must be set."
    exit 1
fi

IFS=',' read -r -a PLATFORM_LIST <<< "$PLATFORMS"

# Remove any existing manifest and image
podman manifest rm "${OPERATOR_IMAGE_TAGGED}" 2>/dev/null || true
podman rmi "${OPERATOR_IMAGE_TAGGED}" 2>/dev/null || true

podman manifest create "${OPERATOR_IMAGE_TAGGED}"

for platform in "${PLATFORM_LIST[@]}"; do
    podman build \
        --build-arg BUILD_ARCH="$ARCH" \
        --platform "$platform" \
        --manifest "${OPERATOR_IMAGE_TAGGED}" \
        -f build/operator/Dockerfile \
        .
done
