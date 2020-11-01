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
source ${SCRIPTS_PATH}/kubevirtci.sh
kubevirtci::install

$(kubevirtci::path)/cluster-up/up.sh

if [[ "$KUBEVIRT_PROVIDER" =~ (ocp|okd)- ]]; then
    echo 'Remove components we do not need to save some resources'
    ./cluster/kubectl.sh delete ns openshift-monitoring --wait=false
    ./cluster/kubectl.sh delete ns openshift-marketplace --wait=false
    ./cluster/kubectl.sh delete ns openshift-cluster-samples-operator --wait=false
fi

if [[ "$KUBEVIRT_PROVIDER" =~ k8s- ]]; then
    echo 'Install NetworkManager on nodes'
    for node in $(./cluster/kubectl.sh get nodes --no-headers | awk '{print $1}'); do
        ./cluster/cli.sh ssh ${node} sudo -- yum install -y yum-plugin-copr
        ./cluster/cli.sh ssh ${node} sudo -- yum copr enable -y networkmanager/NetworkManager-1.22
        ./cluster/cli.sh ssh ${node} sudo -- yum install -y NetworkManager NetworkManager-ovs
        ./cluster/cli.sh ssh ${node} sudo -- systemctl daemon-reload
        ./cluster/cli.sh ssh ${node} sudo -- systemctl restart NetworkManager
        echo "Check NetworkManager is working fine on node $node"
        ./cluster/cli.sh ssh ${node} -- nmcli device show > /dev/null
    done
fi

if [[ "$KUBEVIRT_PROVIDER" =~ k8s- ]]; then
    echo 'Installing Open vSwitch on nodes'
    for node in $(./cluster/kubectl.sh get nodes --no-headers | awk '{print $1}'); do
        ./cluster/cli.sh ssh ${node} -- sudo yum install -y http://cbs.centos.org/kojifiles/packages/openvswitch/2.9.2/1.el7/x86_64/openvswitch-2.9.2-1.el7.x86_64.rpm http://cbs.centos.org/kojifiles/packages/openvswitch/2.9.2/1.el7/x86_64/openvswitch-devel-2.9.2-1.el7.x86_64.rpm http://cbs.centos.org/kojifiles/packages/dpdk/17.11/3.el7/x86_64/dpdk-17.11-3.el7.x86_64.rpm
        ./cluster/cli.sh ssh ${node} -- sudo systemctl daemon-reload
        ./cluster/cli.sh ssh ${node} -- sudo systemctl restart openvswitch
    done
fi
