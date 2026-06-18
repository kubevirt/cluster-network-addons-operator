#!/bin/bash
#
# Copyright 2024 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -xeo pipefail

source hack/components/docker-utils.sh

export OCI_BIN=${OCI_BIN:-$(docker-utils::determine_cri_bin)}

# Pre-pull the kubevirtci cluster image with retry logic to avoid failures during setup
function cluster-utils::prepull_cluster_image() {
  local cluster_image="quay.io/kubevirtci/${KUBEVIRT_PROVIDER}:${KUBEVIRTCI_TAG}"
  local max_attempts=3
  local attempt=1
  local wait_time=5

  echo "Pre-pulling kubevirtci cluster image: ${cluster_image}"

  while [ $attempt -le $max_attempts ]; do
    echo "Attempt $attempt of $max_attempts to pull ${cluster_image}"
    if ${OCI_BIN} pull ${cluster_image}; then
      echo "Successfully pulled ${cluster_image}"
      return 0
    else
      echo "Failed to pull ${cluster_image} on attempt $attempt"
      if [ $attempt -lt $max_attempts ]; then
        echo "Waiting ${wait_time} seconds before retry..."
        sleep $wait_time
        wait_time=$((wait_time * 2))  # Exponential backoff
      fi
      attempt=$((attempt + 1))
    fi
  done

  echo "WARNING: Failed to pull ${cluster_image} after $max_attempts attempts"
  echo "This may be expected if the image doesn't support the current architecture"
  echo "The kubevirtci cluster-up script will handle image pulling as needed"
  return 0
}
