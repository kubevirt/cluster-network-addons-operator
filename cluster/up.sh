#!/bin/bash
#
# Copyright 2018-2019 Red Hat, Inc.
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

set -ex

SCRIPTS_PATH="$(dirname "$(realpath "$0")")"
source ${SCRIPTS_PATH}/cluster.sh
source ${SCRIPTS_PATH}/../hack/components/docker-utils.sh

export KUBEVIRT_DEPLOY_PROMETHEUS=true
export KUBEVIRT_DEPLOY_PROMETHEUS_ALERTMANAGER=true
export KUBEVIRT_DEPLOY_GRAFANA=true

cluster::install

# Pre-pull kubevirtci cluster image to avoid timeout issues during cluster-up
# The cluster-up script downloads this large image, which can be slow in CI environments
# and cause Prow job timeouts. Pre-pulling ensures the image is cached.
OCI_BIN=${OCI_BIN:-$(docker-utils::determine_cri_bin)}
CLUSTER_IMAGE="quay.io/kubevirtci/${KUBEVIRT_PROVIDER}:${KUBEVIRTCI_TAG}"
echo "Pre-pulling kubevirtci cluster image: ${CLUSTER_IMAGE}..."
${OCI_BIN} pull ${CLUSTER_IMAGE} || true

$(cluster::path)/cluster-up/up.sh

if [[ "$KUBEVIRT_PROVIDER" =~ k8s- ]]; then
    echo 'Enable Open vSwitch'
    for node in $(./cluster/kubectl.sh get nodes --no-headers -o=custom-columns=NAME:.metadata.name); do
        ./cluster/cli.sh ssh ${node} -- sudo systemctl enable --now openvswitch
    done
fi
