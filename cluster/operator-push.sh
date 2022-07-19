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

if [[ "${KUBEVIRT_PROVIDER}" == external ]]; then
    if [[ ! -v DEV_IMAGE_REGISTRY ]]; then
        echo "Missing DEV_IMAGE_REGISTRY variable"
        return 1
    fi
    manifest_registry=$DEV_IMAGE_REGISTRY
    push_registry=$manifest_registry
else
    manifest_registry=registry:5000
    push_registry=127.0.0.1:$(./cluster/cli.sh ports registry | tr -d '\r')
fi

# Cleanup previously generated manifests
rm -rf _out/
# Copy release manifests as a base for generated ones, this should make it possible to upgrade
cp -r manifests/ _out/
IMAGE_REGISTRY=$manifest_registry DEPLOY_DIR=_out make gen-manifests

make cluster-clean

IMAGE_REGISTRY=$push_registry make docker-build-operator docker-push-operator

./cluster/kubectl.sh create -f _out/cluster-network-addons/${VERSION}/namespace.yaml
./cluster/kubectl.sh create -f _out/cluster-network-addons/${VERSION}/network-addons-config.crd.yaml
