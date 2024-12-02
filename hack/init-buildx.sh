#!/bin/bash

check_buildx() {
  export DOCKER_CLI_EXPERIMENTAL=enabled

  if ! docker buildx > /dev/null 2>&1; then
     mkdir -p ~/.docker/cli-plugins
     BUILDX_VERSION=$(curl -s https://api.github.com/repos/docker/buildx/releases/latest | jq -r .tag_name)
     ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
     curl -L https://github.com/docker/buildx/releases/download/"${BUILDX_VERSION}"/buildx-"${BUILDX_VERSION}".linux-"${ARCH}" --output ~/.docker/cli-plugins/docker-buildx
     chmod a+x ~/.docker/cli-plugins/docker-buildx
  fi
}

create_or_use_buildx_builder() {
  local builder_name=$1
  if [ -z "$builder_name" ]; then
    echo "Error: Builder name is required."
    exit 1
  fi

  check_buildx

  current_builder="$(docker buildx inspect "${builder_name}" 2>/dev/null)" || echo "Builder '${builder_name}' not found"

  if ! grep -q "^Driver: docker$" <<<"${current_builder}" && \
     grep -q "linux/amd64" <<<"${current_builder}" && \
     grep -q "linux/arm64" <<<"${current_builder}" && \
     grep -q "linux/s390x" <<<"${current_builder}"; then
    echo "The current builder already has multi-architecture support (amd64, arm64, s390x)."
    echo "Skipping setup as the builder is already configured correctly."
    exit 0
  fi

  # Check if the builder already exists by parsing the output of `docker buildx ls`
  # We check if the builder_name appears in the list of active builders
  existing_builder=$(docker buildx ls | grep -w "$builder_name" | awk '{print $1}')

  if [ -n "$existing_builder" ]; then
    echo "Builder '$builder_name' already exists."
    echo "Using existing builder '$builder_name'."
    docker buildx use "$builder_name"
  else
    echo "Creating a new Docker Buildx builder: $builder_name"
    docker buildx create --driver-opt network=host --use --name "$builder_name"
    echo "The new builder '$builder_name' has been created and set as active."
  fi
}

if [ $# -eq 1 ]; then
  create_or_use_buildx_builder "$1"
else
  echo "Usage: $0 <builder_name>"
  echo "Example: $0 mybuilder"
  exit 1
fi
