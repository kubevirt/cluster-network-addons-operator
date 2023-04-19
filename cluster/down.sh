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
cluster::install

echo
echo
echo
./cluster/ssh.sh node01 "sudo ./audit2rbac -f /var/log/k8s-audit/k8s-audit.log --serviceaccount cluster-network-addons:cluster-network-addons-operator" || true
echo
echo
echo
./cluster/ssh.sh node01 "sudo ./audit2rbac -f /var/log/k8s-audit/k8s-audit.log --serviceaccount cluster-network-addons:default" || true
echo
echo
echo

$(cluster::path)/cluster-up/down.sh
