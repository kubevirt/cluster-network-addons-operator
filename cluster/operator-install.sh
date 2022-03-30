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

./cluster/kubectl.sh create -f _out/cluster-network-addons/${VERSION}/operator.yaml
if [[ ! $(./cluster/kubectl.sh -n cluster-network-addons wait deployment cluster-network-addons-operator --for condition=Available --timeout=600s) ]]; then
	echo "Failed to wait for CNAO deployment to be ready"
	./cluster/kubectl.sh get pods -n cluster-network-addons
	./cluster/kubectl.sh describe deployment cluster-network-addons-operator -n cluster-network-addons
	exit 1
fi
