#!/usr/bin/env bash

set -xo pipefail

function docker-utils::get_image_digest() {
  echo "${2}@$(docker run --rm quay.io/skopeo/stable:latest inspect "docker://${1}" | jq -r '.Digest')"
}

function docker-utils::check_image_exists() {
  docker run --rm quay.io/skopeo/stable:latest list-tags "docker://${1}" | grep "\"${2}\""
}