#!/usr/bin/env bash

set -xo pipefail

# The get_image_digest function retrieves the image digest from the registry, according to the
# image name (registry:tag format).
#
# Parameters:
# 1. image name in registry:tag format
# 2. image name without the tag
#
# Returns
# The image digest
#
function docker-utils::get_image_digest() {
  echo "${2}@$(docker run --rm --security-opt seccomp=unconfined quay.io/skopeo/stable:latest inspect "docker://${1}" | jq -r '.Digest')"
}

# The check_image_exists function checks if an image already exists in the registry.
#
# Parameters:
# 1. image name in registry:tag format
# 2. image tag
#
# returns the image tag if found; else, returns an empty result
#
function docker-utils::check_image_exists() {
  docker run --rm --security-opt seccomp=unconfined quay.io/skopeo/stable:latest list-tags "docker://${1}" | grep "\"${2}\""
}
